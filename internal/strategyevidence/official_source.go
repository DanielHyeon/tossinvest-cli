package strategyevidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type OfficialCollectionRequest struct {
	OperationID       string
	EntityID          string
	PageLimit         int
	ResponseByteLimit int64
}

type OfficialRecord struct {
	Authority          SourceAuthority `json:"authority"`
	IssuerIdentity     string          `json:"issuer_identity"`
	Symbol             string          `json:"symbol"`
	SourceRecordID     string          `json:"source_record_id"`
	RevisionIdentity   string          `json:"revision_identity"`
	Form               string          `json:"form"`
	FiledDate          string          `json:"filed_date"`
	AcceptanceAt       time.Time       `json:"acceptance_at,omitempty"`
	SourceAvailableAt  time.Time       `json:"source_available_at"`
	PrimaryDocument    string          `json:"primary_document,omitempty"`
	ReportName         string          `json:"report_name,omitempty"`
	FilerName          string          `json:"filer_name,omitempty"`
	CorrectionNotation string          `json:"correction_notation,omitempty"`
}

type OfficialBatch struct {
	Authority     SourceAuthority  `json:"authority"`
	ContractID    string           `json:"contract_id"`
	SchemaVersion string           `json:"schema_version"`
	Pages         int              `json:"pages"`
	Records       []OfficialRecord `json:"records"`
	Digest        string           `json:"digest"`
	Complete      bool             `json:"complete"`
}

type OfficialBatchSink interface {
	Commit(context.Context, OfficialBatch) error
}

type OfficialAdapter struct {
	policy SourcePolicy
	base   *Adapter
}

func NewOfficialAdapter(policy SourcePolicy, transport Transport, credentials CredentialProvider, budget *SharedRateBudget) (*OfficialAdapter, error) {
	if err := policy.validateOfficial(); err != nil {
		return nil, err
	}
	if transport == nil || budget == nil {
		return nil, ErrSourceDisabled
	}
	if policy.CredentialRequired && credentials == nil {
		return nil, ErrSourceCredential
	}
	base := NewAdapter(policy, transport, credentials)
	base.shared = budget
	base.budgetKey = string(policy.Authority) + "/" + policy.ContractID
	return &OfficialAdapter{policy: policy, base: base}, nil
}

func (a *OfficialAdapter) CollectAndCommit(ctx context.Context, request OfficialCollectionRequest, sink OfficialBatchSink) (OfficialBatch, error) {
	batch, err := a.Collect(ctx, request)
	if err != nil {
		return OfficialBatch{}, err
	}
	if !batch.Complete || sink == nil {
		return OfficialBatch{}, ErrSourceIncomplete
	}
	if err := sink.Commit(ctx, batch); err != nil {
		return OfficialBatch{}, err
	}
	return batch, nil
}

func (a *OfficialAdapter) Collect(ctx context.Context, request OfficialCollectionRequest) (OfficialBatch, error) {
	request.OperationID = strings.TrimSpace(request.OperationID)
	request.EntityID = strings.TrimSpace(request.EntityID)
	if request.OperationID == "" || request.EntityID == "" {
		return OfficialBatch{}, ErrSourceBoundExceeded
	}
	if request.PageLimit == 0 {
		request.PageLimit = a.policy.MaxPages
	}
	if request.ResponseByteLimit == 0 {
		request.ResponseByteLimit = a.policy.MaxResponseBytes
	}
	if request.PageLimit <= 0 || request.PageLimit > a.policy.MaxPages || request.ResponseByteLimit <= 0 || request.ResponseByteLimit > a.policy.MaxResponseBytes {
		return OfficialBatch{}, ErrSourceBoundExceeded
	}
	operationCtx, cancel := context.WithTimeout(ctx, a.policy.OperationDeadline)
	defer cancel()
	switch a.policy.Authority {
	case AuthoritySEC:
		return a.collectSEC(operationCtx, request)
	case AuthorityOpenDART:
		return a.collectDART(operationCtx, request)
	default:
		return OfficialBatch{}, ErrSourceUnavailable
	}
}

func (a *OfficialAdapter) fetchPage(ctx context.Context, request OfficialCollectionRequest, page int, resource string, remainingBytes int64) (FetchResult, error) {
	if remainingBytes <= 0 {
		return FetchResult{}, ErrSourceBoundExceeded
	}
	operationDeadline := a.policy.OperationDeadline
	if deadline, ok := ctx.Deadline(); ok {
		operationDeadline = time.Until(deadline)
		if operationDeadline <= 0 {
			return FetchResult{}, context.DeadlineExceeded
		}
	}
	requestDeadline := a.policy.RequestDeadline
	if operationDeadline < requestDeadline {
		requestDeadline = operationDeadline
	}
	return a.base.Fetch(ctx, FetchRequest{
		OperationID: request.OperationID, PageLimit: request.PageLimit, ResponseByteLimit: remainingBytes,
		Concurrency: 1, RequestDeadline: requestDeadline, OperationDeadline: operationDeadline,
		Page: page, PageSize: a.policy.PageSize, Resource: resource,
	})
}

var digitsOnly = regexp.MustCompile(`^[0-9]+$`)

type secRecent struct {
	AccessionNumber    []string `json:"accessionNumber"`
	FilingDate         []string `json:"filingDate"`
	ReportDate         []string `json:"reportDate"`
	AcceptanceDateTime []string `json:"acceptanceDateTime"`
	Form               []string `json:"form"`
	PrimaryDocument    []string `json:"primaryDocument"`
}

type secSubmissions struct {
	CIK       string   `json:"cik"`
	Name      string   `json:"name"`
	Tickers   []string `json:"tickers"`
	Exchanges []string `json:"exchanges"`
	Filings   struct {
		Recent secRecent `json:"recent"`
		Files  []struct {
			Name        string `json:"name"`
			FilingCount int    `json:"filingCount"`
			FilingFrom  string `json:"filingFrom"`
			FilingTo    string `json:"filingTo"`
		} `json:"files"`
	} `json:"filings"`
}

func (a *OfficialAdapter) collectSEC(ctx context.Context, request OfficialCollectionRequest) (OfficialBatch, error) {
	if len(request.EntityID) != 10 || !digitsOnly.MatchString(request.EntityID) {
		return OfficialBatch{}, ErrSourceSchemaDrift
	}
	first, err := a.fetchPage(ctx, request, 1, "CIK"+request.EntityID+".json", request.ResponseByteLimit)
	if err != nil {
		return OfficialBatch{}, err
	}
	var root secSubmissions
	if err := decodeNoDuplicateJSON(first.Body, &root); err != nil || root.CIK != request.EntityID || strings.TrimSpace(root.Name) == "" || len(root.Tickers) == 0 || len(root.Exchanges) == 0 {
		return OfficialBatch{}, ErrSourceSchemaDrift
	}
	records, err := secRecords(root.CIK, root.Tickers[0], first, root.Filings.Recent)
	if err != nil {
		return OfficialBatch{}, err
	}
	pagesRequired := 1 + len(root.Filings.Files)
	if pagesRequired > request.PageLimit || pagesRequired > a.policy.MaxPages {
		return OfficialBatch{}, ErrSourceIncomplete
	}
	totalBytes := int64(len(first.Body))
	for index, file := range root.Filings.Files {
		if strings.TrimSpace(file.Name) == "" || file.FilingCount < 0 {
			return OfficialBatch{}, ErrSourceSchemaDrift
		}
		page, fetchErr := a.fetchPage(ctx, request, index+2, file.Name, request.ResponseByteLimit-totalBytes)
		if fetchErr != nil {
			return OfficialBatch{}, fetchErr
		}
		totalBytes += int64(len(page.Body))
		if totalBytes > request.ResponseByteLimit {
			return OfficialBatch{}, ErrSourceBoundExceeded
		}
		var historical secRecent
		if err := decodeNoDuplicateJSON(page.Body, &historical); err != nil {
			return OfficialBatch{}, ErrSourceSchemaDrift
		}
		parsed, parseErr := secRecords(root.CIK, root.Tickers[0], page, historical)
		if parseErr != nil || (file.FilingCount > 0 && len(parsed) != file.FilingCount) {
			return OfficialBatch{}, ErrSourceSchemaDrift
		}
		records = append(records, parsed...)
	}
	return completeOfficialBatch(a.policy, pagesRequired, records)
}

func secRecords(cik, symbol string, result FetchResult, recent secRecent) ([]OfficialRecord, error) {
	length := len(recent.AccessionNumber)
	if length == 0 || len(recent.FilingDate) != length || len(recent.ReportDate) != length || len(recent.AcceptanceDateTime) != length || len(recent.Form) != length || len(recent.PrimaryDocument) != length {
		return nil, ErrSourceSchemaDrift
	}
	observed := resultObservedAt(result)
	if observed.IsZero() {
		return nil, ErrSourceSchemaDrift
	}
	records := make([]OfficialRecord, 0, length)
	for index := 0; index < length; index++ {
		accepted, err := time.Parse(time.RFC3339Nano, recent.AcceptanceDateTime[index])
		if err != nil || !validISODate(recent.FilingDate[index]) || strings.TrimSpace(recent.AccessionNumber[index]) == "" || strings.TrimSpace(recent.Form[index]) == "" {
			return nil, ErrSourceSchemaDrift
		}
		records = append(records, OfficialRecord{
			Authority: AuthoritySEC, IssuerIdentity: cik, Symbol: strings.ToUpper(strings.TrimSpace(symbol)),
			SourceRecordID: recent.AccessionNumber[index], RevisionIdentity: recent.AccessionNumber[index],
			Form: recent.Form[index], FiledDate: recent.FilingDate[index], AcceptanceAt: accepted.UTC(),
			SourceAvailableAt: observed, PrimaryDocument: recent.PrimaryDocument[index],
		})
	}
	return records, nil
}

type dartListResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	PageNo     int    `json:"page_no"`
	PageCount  int    `json:"page_count"`
	TotalCount int    `json:"total_count"`
	TotalPage  int    `json:"total_page"`
	List       []struct {
		CorpClass   string `json:"corp_cls"`
		CorpName    string `json:"corp_name"`
		CorpCode    string `json:"corp_code"`
		StockCode   string `json:"stock_code"`
		ReportName  string `json:"report_nm"`
		ReceiptNo   string `json:"rcept_no"`
		FilerName   string `json:"flr_nm"`
		ReceiptDate string `json:"rcept_dt"`
		Remark      string `json:"rm"`
	} `json:"list"`
}

func (a *OfficialAdapter) collectDART(ctx context.Context, request OfficialCollectionRequest) (OfficialBatch, error) {
	if len(request.EntityID) != 8 || !digitsOnly.MatchString(request.EntityID) {
		return OfficialBatch{}, ErrSourceSchemaDrift
	}
	var records []OfficialRecord
	var totalBytes int64
	totalPages := 1
	totalCount := -1
	for pageNo := 1; pageNo <= totalPages; pageNo++ {
		result, err := a.fetchPage(ctx, request, pageNo, "list.json", request.ResponseByteLimit-totalBytes)
		if err != nil {
			return OfficialBatch{}, err
		}
		totalBytes += int64(len(result.Body))
		if totalBytes > request.ResponseByteLimit {
			return OfficialBatch{}, ErrSourceBoundExceeded
		}
		var page dartListResponse
		if err := decodeNoDuplicateJSON(result.Body, &page); err != nil {
			return OfficialBatch{}, ErrSourceSchemaDrift
		}
		switch page.Status {
		case "000":
		case "013":
			if pageNo != 1 {
				return OfficialBatch{}, ErrSourceSchemaDrift
			}
			return completeOfficialBatch(a.policy, 1, nil)
		case "010", "011", "012", "901":
			return OfficialBatch{}, ErrSourceCredential
		case "020":
			return OfficialBatch{}, ErrSourceRateLimited
		default:
			return OfficialBatch{}, ErrSourceIncomplete
		}
		if page.PageNo != pageNo || page.PageCount <= 0 || page.PageCount > a.policy.PageSize || page.TotalPage <= 0 || page.TotalCount < 0 || len(page.List) > page.PageCount {
			return OfficialBatch{}, ErrSourceSchemaDrift
		}
		if pageNo == 1 {
			totalPages, totalCount = page.TotalPage, page.TotalCount
			if totalPages > request.PageLimit || totalPages > a.policy.MaxPages {
				return OfficialBatch{}, ErrSourceIncomplete
			}
		} else if page.TotalPage != totalPages || page.TotalCount != totalCount {
			return OfficialBatch{}, ErrSourceSchemaDrift
		}
		observed := resultObservedAt(result)
		if observed.IsZero() {
			return OfficialBatch{}, ErrSourceSchemaDrift
		}
		for _, item := range page.List {
			if item.CorpCode != request.EntityID || len(item.StockCode) != 6 || !digitsOnly.MatchString(item.StockCode) || strings.TrimSpace(item.ReceiptNo) == "" || !validCompactDate(item.ReceiptDate) || strings.TrimSpace(item.ReportName) == "" {
				return OfficialBatch{}, ErrSourceSchemaDrift
			}
			records = append(records, OfficialRecord{
				Authority: AuthorityOpenDART, IssuerIdentity: item.CorpCode, Symbol: item.StockCode,
				SourceRecordID: item.ReceiptNo, RevisionIdentity: item.ReceiptNo, FiledDate: compactToISO(item.ReceiptDate),
				SourceAvailableAt: observed, ReportName: item.ReportName, FilerName: item.FilerName, CorrectionNotation: item.Remark,
			})
		}
	}
	if len(records) != totalCount {
		return OfficialBatch{}, ErrSourceIncomplete
	}
	return completeOfficialBatch(a.policy, totalPages, records)
}

// FetchResult deliberately keeps transport timestamps out of public metadata.
// Official parsing receives it via this package-private carrier.
func resultObservedAt(result FetchResult) time.Time { return result.observedAt }

func completeOfficialBatch(policy SourcePolicy, pages int, records []OfficialRecord) (OfficialBatch, error) {
	batch := OfficialBatch{Authority: policy.Authority, ContractID: policy.ContractID, SchemaVersion: policy.SchemaVersion, Pages: pages, Records: append([]OfficialRecord(nil), records...), Complete: true}
	preimage, err := json.Marshal(struct {
		Authority  SourceAuthority  `json:"authority"`
		ContractID string           `json:"contract_id"`
		Schema     string           `json:"schema"`
		Pages      int              `json:"pages"`
		Records    []OfficialRecord `json:"records"`
	}{batch.Authority, batch.ContractID, batch.SchemaVersion, batch.Pages, batch.Records})
	if err != nil {
		return OfficialBatch{}, err
	}
	sum := sha256.Sum256(preimage)
	batch.Digest = hex.EncodeToString(sum[:])
	return batch, nil
}

func validISODate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validCompactDate(value string) bool {
	_, err := time.Parse("20060102", value)
	return err == nil
}

func compactToISO(value string) string { return value[:4] + "-" + value[4:6] + "-" + value[6:] }

func decodeNoDuplicateJSON(body []byte, target any) error {
	if _, err := canonicalJSON(body); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}
