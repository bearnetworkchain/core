package networkchain

import (
	"github.com/bearnetworkchain/core/ignite/chainconfig"
	"github.com/bearnetworkchain/core/ignite/pkg/checksum"
	"github.com/bearnetworkchain/core/ignite/pkg/confile"
	"github.com/bearnetworkchain/core/ignite/pkg/xfilepath"
)

const (
	SPNCacheDirectory    = "spn"
	BinaryCacheDirectory = "binary-cache"
	BinaryCacheFilename  = "checksums.yml"
)

type BinaryCacheList struct {
	CachedBinaries []Binary `yaml:"cached_binaries"`
}

//二進制將啟動 ID 與構建哈希相關聯，其中構建哈希為 sha256（二進制，源）
type Binary struct {
	LaunchID  uint64
	BuildHash string
}

func (l *BinaryCacheList) Set(launchID uint64, buildHash string) {
	for i, binary := range l.CachedBinaries {
		if binary.LaunchID == launchID {
			l.CachedBinaries[i].BuildHash = buildHash
			return
		}
	}
	l.CachedBinaries = append(l.CachedBinaries, Binary{
		LaunchID:  launchID,
		BuildHash: buildHash,
	})
}

func (l *BinaryCacheList) Get(launchID uint64) (string, bool) {
	for _, binary := range l.CachedBinaries {
		if binary.LaunchID == launchID {
			return binary.BuildHash, true
		}
	}
	return "", false
}

// cacheBinaryForLaunchID 緩存哈希 sha256(sha256(binary) + sourcehash) 以獲取啟動 ID
func cacheBinaryForLaunchID(launchID uint64, binaryHash, sourceHash string) error {
	cachePath, err := getBinaryCacheFilepath()
	if err != nil {
		return err
	}
	var cacheList = BinaryCacheList{}
	err = confile.New(confile.DefaultYAMLEncodingCreator, cachePath).Load(&cacheList)
	if err != nil {
		return err
	}
	cacheList.Set(launchID, checksum.Strings(binaryHash, sourceHash))

	return confile.New(confile.DefaultYAMLEncodingCreator, cachePath).Save(cacheList)
}

// checkBinaryCacheForLaunchID 檢查給定啟動的二進製文件是否已經構建
func checkBinaryCacheForLaunchID(launchID uint64, binaryHash, sourceHash string) (bool, error) {
	cachePath, err := getBinaryCacheFilepath()
	if err != nil {
		return false, err
	}
	var cacheList = BinaryCacheList{}
	err = confile.New(confile.DefaultYAMLEncodingCreator, cachePath).Load(&cacheList)
	if err != nil {
		return false, err
	}
	buildHash, ok := cacheList.Get(launchID)
	return ok && buildHash == checksum.Strings(binaryHash, sourceHash), nil
}

func getBinaryCacheFilepath() (string, error) {
	return xfilepath.Join(
		chainconfig.ConfigDirPath,
		xfilepath.Path(SPNCacheDirectory),
		xfilepath.Path(BinaryCacheDirectory),
		xfilepath.Path(BinaryCacheFilename),
	)()
}
