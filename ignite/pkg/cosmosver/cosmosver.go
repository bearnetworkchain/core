package cosmosver

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
)

// Family 代表 Cosmos-SDK 的系列（命名版本）。
type Family string

const (
	// Launchpad 代表 Cosmos-SDK 的啟動板系列。
	Launchpad Family = "launchpad"

	// Stargate 代表 Cosmos-SDK 的 Stargate 系列。
	Stargate Family = "stargate"
)

const prefix = "v"

// Version 代表一系列 Cosmos SDK 版本。
type Version struct {
	// 版本家族
	Family Family

	// 版本是確切的 sdk 版本字符串。
	Version string

	// 語義是解析的版本。
	Semantic semver.Version
}

var (
	MaxLaunchpadVersion           = newVersion("0.39.99", Launchpad)
	StargateFortyVersion          = newVersion("0.40.0", Stargate)
	StargateFortyFourVersion      = newVersion("0.44.0-alpha", Stargate)
	StargateFortyFiveThreeVersion = newVersion("0.45.3", Stargate)
)

var (
	// Versions 是已知的、已排序的 Cosmos-SDK 版本列表。
	Versions = []Version{
		MaxLaunchpadVersion,
		StargateFortyVersion,
		StargateFortyFourVersion,
	}

	// 最新是 Cosmos-SDK 的最新已知版本。
	Latest = Versions[len(Versions)-1]
)

func newVersion(version string, family Family) Version {
	return Version{
		Family:   family,
		Version:  "v" + version,
		Semantic: semver.MustParse(version),
	}
}

// Parse 解析 Cosmos-SDK 版本。
func Parse(version string) (v Version, err error) {
	v.Version = version

	if v.Semantic, err = semver.Parse(strings.TrimPrefix(version, prefix)); err != nil {
		return v, err
	}

	v.Family = Stargate
	if v.LTE(MaxLaunchpadVersion) {
		v.Family = Launchpad
	}

	return
}

// GTE 檢查 v 是否大於或等於版本。
func (v Version) GTE(version Version) bool {
	return v.Semantic.GTE(version.Semantic)
}

// LT 檢查 v 是否小於版本。
func (v Version) LT(version Version) bool {
	return v.Semantic.LT(version.Semantic)
}

// LTE 檢查 v 是否小於或等於版本。
func (v Version) LTE(version Version) bool {
	return v.Semantic.LTE(version.Semantic)
}

// Is 檢查 v 是否等於版本。
func (v Version) Is(version Version) bool {
	return v.Semantic.EQ(version.Semantic)
}

func (v Version) String() string {
	return fmt.Sprintf("%s - %s", v.Family, v.Version)
}

func (v Version) IsFamily(family Family) bool {
	return v.Family == family
}
