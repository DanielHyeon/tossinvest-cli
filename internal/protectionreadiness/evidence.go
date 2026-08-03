package protectionreadiness

import (
	"crypto/sha256"
	"errors"
)

type observedFileInput struct {
	Bytes        []byte
	ResolvedPath string
	OwnerUID     uint32
	Mode         uint32
	Regular      bool
	Symlink      bool
}

type observedFile struct {
	bytes        []byte
	ResolvedPath string
	OwnerUID     uint32
	Mode         uint32
	Regular      bool
	Symlink      bool
	Size         int64
	seal         [32]byte
}

func newObservedFile(input observedFileInput) (observedFile, error) {
	if len(input.Bytes) == 0 || input.ResolvedPath == "" {
		return observedFile{}, errors.New("protectionreadiness: invalid observed file")
	}
	file := observedFile{bytes: append([]byte(nil), input.Bytes...), ResolvedPath: input.ResolvedPath, OwnerUID: input.OwnerUID, Mode: input.Mode, Regular: input.Regular, Symlink: input.Symlink, Size: int64(len(input.Bytes))}
	file.seal = observedFileSeal(file)
	return file, nil
}

func observedFileSeal(file observedFile) [32]byte {
	hash := sha256.New()
	writeString(hash, string(file.bytes))
	writeString(hash, file.ResolvedPath)
	writeUint64(hash, uint64(file.OwnerUID))
	writeUint64(hash, uint64(file.Mode))
	writeUint64(hash, uint64(file.Size))
	if file.Regular {
		writeString(hash, "regular")
	}
	if file.Symlink {
		writeString(hash, "symlink")
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validObservedFile(file observedFile, policy pinnedTrustPolicy, market Market) bool {
	return file.seal == observedFileSeal(file) && file.Size == int64(len(file.bytes)) && file.Size > 0 && file.Size <= policy.maximumFileBytes &&
		file.ResolvedPath == policy.expectedPaths[market] && file.OwnerUID == policy.requiredOwnerUID && file.Mode == policy.requiredMode && file.Regular && !file.Symlink
}

type supervisorBindingInput struct {
	AccountID       string
	ProfileID       string
	Market          Market
	BuildDigest     string
	ComponentDigest string
	Wired           bool
}

type supervisorBinding struct {
	AccountID       string
	ProfileID       string
	Market          Market
	BuildDigest     string
	ComponentDigest string
	Wired           bool
	seal            [32]byte
}

func newSupervisorBinding(input supervisorBindingInput) (supervisorBinding, error) {
	if input.AccountID == "" || input.ProfileID == "" || !validMarket(input.Market) || !validDigest(input.BuildDigest) || !validDigest(input.ComponentDigest) || !input.Wired {
		return supervisorBinding{}, errors.New("protectionreadiness: invalid supervisor binding")
	}
	binding := supervisorBinding{AccountID: input.AccountID, ProfileID: input.ProfileID, Market: input.Market, BuildDigest: input.BuildDigest, ComponentDigest: input.ComponentDigest, Wired: input.Wired}
	binding.seal = supervisorBindingSeal(binding)
	return binding, nil
}

func supervisorBindingSeal(binding supervisorBinding) [32]byte {
	wired := "unwired"
	if binding.Wired {
		wired = "wired"
	}
	return hashStrings(binding.AccountID, binding.ProfileID, string(binding.Market), binding.BuildDigest, binding.ComponentDigest, wired)
}

func validSupervisorBinding(binding supervisorBinding, scope runtimeScope) bool {
	return binding.Wired && binding.seal == supervisorBindingSeal(binding) && binding.AccountID == scope.AccountID && binding.ProfileID == scope.ProfileID &&
		binding.Market == scope.Market && binding.BuildDigest == scope.BuildDigest
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
