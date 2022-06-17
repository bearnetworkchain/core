package protoanalysis

import (
	"context"
	"os"
	"path/filepath"

	"github.com/emicklei/proto"
	"github.com/pkg/errors"

	"github.com/bearnetworkchain/core/ignite/pkg/localfs"
)

const optionGoPkg = "go_package"

// parser 解析 proto 包。
type parser struct {
	packages []*pkg
}

// parse 解析fs中與pattern匹配的proto文件並返回
// proto 包的低級表示。
func parse(ctx context.Context, path, pattern string) ([]*pkg, error) {
	pr := &parser{}

	paths, err := localfs.Search(path, pattern)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err := pr.parseFile(path); err != nil {
			return nil, errors.Wrapf(err, "file: %s", path)
		}
	}

	return pr.packages, nil
}

// pkg 代表一個原型包。
type pkg struct {
	// 原型包的名稱。
	name string

	// fs 中 proto 包的目錄。
	dir string

	// files 是構造 proto 包的 proto 文件列表。
	files []file
}

// file 表示已解析的 proto 文件。
type file struct {
	// fs 中 proto 文件的路徑。
	path string

	// 解析的數據。
	pkg      *proto.Package
	imports  []string // imported protos.
	options  []*proto.Option
	messages []*proto.Message
	services []*proto.Service
}

func (p *pkg) options() (o []*proto.Option) {
	for _, f := range p.files {
		o = append(o, f.options...)
	}

	return
}

func (p *pkg) messages() (m []*proto.Message) {
	for _, f := range p.files {
		m = append(m, f.messages...)
	}

	return
}

func (p *pkg) services() (s []*proto.Service) {
	for _, f := range p.files {
		s = append(s, f.services...)
	}

	return
}

func (p *parser) parseFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	def, err := proto.NewParser(f).Parse()
	if err != nil {
		return err
	}

	var pkgName string

	proto.Walk(
		def,
		proto.WithPackage(func(p *proto.Package) { pkgName = p.Name }),
	)

	var pp *pkg
	for _, v := range p.packages {
		if pkgName == v.name {
			pp = v
			break
		}
	}
	if pp == nil {
		pp = &pkg{
			name: pkgName,
			dir:  filepath.Dir(path),
		}
		p.packages = append(p.packages, pp)
	}

	pf := file{
		path: path,
	}

	proto.Walk(
		def,
		proto.WithPackage(func(p *proto.Package) { pf.pkg = p }),
		proto.WithImport(func(s *proto.Import) { pf.imports = append(pf.imports, s.Filename) }),
		proto.WithOption(func(o *proto.Option) { pf.options = append(pf.options, o) }),
		proto.WithMessage(func(m *proto.Message) { pf.messages = append(pf.messages, m) }),
		proto.WithService(func(s *proto.Service) { pf.services = append(pf.services, s) }),
	)

	pp.files = append(pp.files, pf)

	return nil
}
