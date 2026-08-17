//go:build !unix

package main

import "errors"

func openReceipt(string) (receiptStore, error) {
	return nil, errors.New("A112 M-B1 secure receipt collection is unsupported on this platform")
}
