package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"gopkg.in/yaml.v2"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	workDir string
	idlDir string
	idlWorkDir string
	srvDir string
	re_import *regexp.Regexp = regexp.MustCompile(`(?sm)import\s+\"([\w\./_\-]*)\";`)
)

type IDLFolder struct {
	prod string
	srv string
	ProdMap    map[string]*ProdIDLFolder
	conf    *Config
	fileMap map[string]struct{}

	Files map[string]*IDLFile
}

type ProdIDLFolder struct {
	Name string
	Files map[string]*IDLFile
	SubSysMap map[string]*SubSysIDLFolder
}

func (p *ProdIDLFolder) compile() {
	defer func() {
		for _, sys := range p.SubSysMap {
			sys.compile(p.Name)
		}
	}()

	if len(p.Files) == 0 {
		return
	}

	os.Chdir(filepath.Join(idlWorkDir, p.Name))

	var m string
	mMap := map[string]string{}

	for _, f := range p.Files {
		for _, imp := range f.Imps {
			mMap[imp.Orig] = imp.Replace
		}
	}

	for orig, rep := range mMap {
		m += fmt.Sprintf("M%s=%s,", orig, rep)
	}
	m = strings.TrimRight(m, ",")
	fmt.Println("m :", m)

	args := []string{
		"-I.",
		"-I"+idlWorkDir,
		"--go_out=" + m + ":.",
	}

	for f, _ := range p.Files {
		args = append(args, f)
	}
	cmd := exec.Command("protoc", args...)
	fmt.Println(cmd.Args)


	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
		os.Exit(1)
	}
}



type SubSysIDLFolder struct {
	Name string
	Files map[string]*IDLFile
	ModMap map[string]*ModIDLFolder
}


func (s *SubSysIDLFolder) compile(prod string) {
	defer func() {
		for _, mod := range s.ModMap {
			fmt.Println("prod:", prod, "sys:", s.Name)

			mod.compile(prod, s.Name)
		}
	}()

	if len(s.Files) == 0 {
		return
	}

	os.Chdir(filepath.Join(idlWorkDir, prod, s.Name))

	var m string
	mMap := map[string]string{}

	for _, f := range s.Files {
		for _, imp := range f.Imps {
			mMap[imp.Orig] = imp.Replace
		}
	}

	for orig, rep := range mMap {
		m += fmt.Sprintf("M%s=%s,", orig, rep)
	}
	m = strings.TrimRight(m, ",")
	fmt.Println("m :", m)

	args := []string{
		"-I.",
		"-I"+idlWorkDir,
		"--go_out=" + m + ":.",
	}

	for f, _ := range s.Files {
		args = append(args, f)
	}
	cmd := exec.Command("protoc", args...)
	fmt.Println(cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
		os.Exit(1)
	}
}



type ModIDLFolder struct {
	Name string
	Files map[string]*IDLFile
}


func (mod *ModIDLFolder) compile(prod, sys string) {
	if len(mod.Files) == 0 {
		return
	}

	fmt.Println("cur dir :", filepath.Join(idlWorkDir, prod, sys, mod.Name))
	os.Chdir(filepath.Join(idlWorkDir, prod, sys, mod.Name))

	var m string
	mMap := map[string]string{}

	for _, f := range mod.Files {
		for _, imp := range f.Imps {
			mMap[imp.Orig] = imp.Replace
		}
	}

	for orig, rep := range mMap {
		m += fmt.Sprintf("M%s=%s,", orig, rep)
	}
	m = strings.TrimRight(m, ",")
	fmt.Println("m :", m)

	args := []string{
		"-I.",
		"-I"+idlWorkDir,
		"--go_out=" + m + ":.",
		"--micro_out=" + m + ":.",
	}

	for f, _ := range mod.Files {
		args = append(args, f)
	}
	cmd := exec.Command("protoc", args...)
	fmt.Println(cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
		os.Exit(1)
	}
}

type IDLFile struct {
	Name string
	Imps map[string]*IDLImportDesc
}

type IDLImportDesc struct {
	Name    string
	Orig    string
	Replace string
}

func (idl *IDLFolder) snapshot() {
	fmt.Println("prod:", idl.prod)
	fmt.Println("srv:", idl.srv)
	fmt.Printf("conf:%+v\n", idl.conf)

	fmt.Println("file list:")
	for f, _ := range idl.fileMap {
		fmt.Printf("\t%s\n", f)
	}

	fmt.Println("")
	fmt.Println("prod list:")
	for pname, prod := range idl.ProdMap {
		fmt.Println("prod:", pname)

		if len(prod.Files) > 0 {
			fmt.Printf("\tfile list:\n")
			for fname, f := range prod.Files {
				fmt.Printf("\t\t%s\n", fname)
				for _, imp := range f.Imps {
					fmt.Printf("\t\t\torig:%s, replace:%s\n", imp.Orig, imp.Replace)
				}
			}
		}

		if len(prod.SubSysMap) > 0 {
			fmt.Printf("\tsubsys list:\n")
			for s, sys := range prod.SubSysMap {
				fmt.Printf("\t\t%s\n", s)
				if len(sys.Files) > 0 {
					fmt.Printf("\t\tfile list:\n")
					for fname, f := range sys.Files {
						fmt.Printf("\t\t\t%s\n", fname)
						for _, imp := range f.Imps {
							fmt.Printf("\t\t\t\torig:%s, replace:%s\n",  imp.Orig, imp.Replace)
						}
					}
				}

				if len(sys.ModMap) > 0 {
					fmt.Printf("\t\tmod list:\n")
					for mname, mod := range sys.ModMap {
						fmt.Printf("\t\t\t%s\n", mname)
						if len(mod.Files) > 0 {
							fmt.Printf("\t\t\t\tfile list:\n")
							for fname, f := range mod.Files {
								fmt.Printf("\t\t\t\t%s\n", fname)
								for _, imp := range f.Imps {
									fmt.Printf("\t\t\t\t\torig:%s, replace:%s\n", imp.Orig, imp.Replace)
								}
							}
						}
					}
				}
			}
		}

	}

}

func (idl *IDLFolder) Extract() {
	os.Chdir(idlDir)

	for _, dep := range idl.conf.Depends {
		ss := strings.Split(dep, "/")
		if len(ss) != 4 {
			fmt.Println("dep must meet the rules that product/subsys/module/proto")
			os.Exit(1)
		}

		idl.extract(dep)
	}

	idl.snapshot()

	idl.copyFile()

	idl.compile()
}

func (idl *IDLFolder) Cleanup() {
	os.Chdir(idlWorkDir)

	filepath.Walk(idlWorkDir, RemoveProtoFile)
}

func (idl *IDLFolder) Transfer() {
	os.Chdir(srvDir)
	os.RemoveAll("idl")

	//fmt.Println("idlwork:", idlWorkDir)
	os.MkdirAll("idl", 0777)
	copyDirectory(idlWorkDir, "idl")
}

func (idl *IDLFolder) compile() {
	os.Chdir(idlWorkDir)

	args := []string{
		"-I.",
		"--go_out=.",
	}

	for f, _ := range idl.Files {
		args = append(args, f)
	}
	cmd := exec.Command("protoc", args...)
	fmt.Println(cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
		os.Exit(1)
	}

	for _, prod := range idl.ProdMap {
		prod.compile()
	}
}

func (idl *IDLFolder) copyFile() {

	for f, _ := range idl.fileMap {

		src := filepath.Join(idlDir, filepath.FromSlash(f))
		dst := filepath.Join(idlWorkDir, filepath.FromSlash(f))
		//dstDir := filepath.Dir(dst)
		os.MkdirAll(filepath.Dir(dst), 0777)
		fmt.Println(src, dst)
		if err := copyFile(src, dst); err != nil {
			fmt.Println("copy file error:", err)
			os.Exit(1)
		}
	}


}


func copyDirectory(src, dst string) {
	files, _ := ioutil.ReadDir(src)
	for _, file := range files {
		//fmt.Println("file:", file.Name())
		subSrc := filepath.Join(src, file.Name())
		subDst := filepath.Join(dst, file.Name())
		if file.IsDir() {
			os.MkdirAll(subDst, 0777)
			copyDirectory(subSrc, subDst)
		} else {
			copyFile(subSrc, subDst)
		}
	}
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func (idl *IDLFolder) extract(in string) {
	if _, ok := idl.fileMap[in]; ok {
		return
	}

	ss := strings.Split(in, "/")
	depth := len(ss)

	if depth == 0  {
		return
	}

	idl.fileMap[in] = struct{}{}

	if depth == 1 {
		f := &IDLFile{
			Name:  in,
			Imps: nil,
		}
		idl.Files[in] = f
		idl.getIDLFile(in, f)
		return
	}

	var ok bool
	var prod *ProdIDLFolder

	if depth > 1 {
		if prod, ok = idl.ProdMap[ss[0]]; !ok {
			prod = &ProdIDLFolder{
				SubSysMap: map[string]*SubSysIDLFolder{},
				Files: map[string]*IDLFile{},
				Name:ss[0],
			}

			idl.ProdMap[ss[0]] = prod
		}
	}

	var subsys *SubSysIDLFolder
	if depth >= 2 {
		if subsys, ok = prod.SubSysMap[ss[1]]; !ok {
			subsys = &SubSysIDLFolder{
				ModMap: map[string]*ModIDLFolder{},
				Files: map[string]*IDLFile{},
				Name: ss[1],
			}
			prod.SubSysMap[ss[1]] = subsys
		}

		if depth == 2 {
			return
		}
	}

	var mod *ModIDLFolder
	if depth >= 3 {
		if mod, ok = subsys.ModMap[ss[2]]; !ok {
			mod = &ModIDLFolder{
				Files: map[string]*IDLFile{},
				Name: ss[2],
			}
			subsys.ModMap[ss[2]] = mod
		}

		if depth == 3 {
			return
		}
	}

	if depth >= 4 {
		if _, ok = mod.Files[ss[3]]; !ok {
			f := &IDLFile{
				Name:  ss[3],
				Imps: map[string]*IDLImportDesc{},
			}
			mod.Files[ss[3]] = f
			idl.getIDLFile(in, f)
		}
	}
}

func (idl *IDLFolder) getIDLFile(in string, out *IDLFile)  {
	abs := filepath.Join(idlDir, filepath.ToSlash(in))

	body, err := ioutil.ReadFile(abs)
	if err != nil {
		fmt.Println("read file error:", err)
		os.Exit(1)
	}

	dyProd := "dy-" + idl.prod

	impss := re_import.FindAllStringSubmatch(string(body), -1)
	for _, imps := range impss {
		fmt.Println("imps:", imps[1])
		ss := strings.Split(imps[1], "/")

		slen := len(ss)

		if slen == 0 {
			fmt.Println("import proto cann't be empty: ", in)
			os.Exit(1)
		}
		var rep string
		if slen == 1 {
			rep = strings.Join([]string{"github.com", dyProd, srv, "idl"}, "/")
		} else if slen > 1 {
			ssTmp := append([]string{"github.com", dyProd, srv, "idl"}, ss[:slen-1]...)
			rep = strings.Join(ssTmp, "/")
		}

		desc := &IDLImportDesc{
			Name:    imps[1],
			Orig:    imps[1],
			Replace: rep,
		}

		out.Imps[imps[1]] = desc

		// 递归继续查询
		idl.extract(imps[1])
	}
}

func (idl *IDLFolder) LoadConfig(conf string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	abs := filepath.Join(wd, conf)
	body, err := ioutil.ReadFile(abs)
	if err != nil {
		fmt.Printf("read file '%s' error: %v\n", conf, err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(body, idl.conf)
	if err != nil {
		fmt.Printf("unmarshal config '%s' error: %v\n", conf, err)
		os.Exit(1)
	}
}

func (idl *IDLFolder) PrepareEnv() {
	u, err := user.Current()
	if err != nil {
		fmt.Println("get current user error:", err)
		os.Exit(1)
	}

	home := u.HomeDir
	workDir = filepath.Join(home, ".dy")

	os.MkdirAll(workDir, 0777)
	// 下载idl仓库
	idlDir = filepath.Join(workDir, "idl")
	_, err = os.Stat(idlDir)
	if err != nil {
		if os.IsNotExist(err) {
			cloneIDL()
		} else {
			fmt.Println("stat idl dir error:", err)
			os.Exit(1)
		}
	} else {
		if err1 := pullIDL(); err1 != nil {
			os.RemoveAll(idlDir)
			cloneIDL()
		}
	}

	// 创建 idl-work
	idlWorkDir = filepath.Join(workDir, "idl-work")
	os.MkdirAll(idlWorkDir, 0777)

	//os.MkdirAll(filepath.Join(idlWorkDir, "idl"), 0777)
}

func cloneIDL() {
	os.Chdir(workDir)

	cmd := exec.Command("git", "clone", "https://github.com/dy-global/idl.git")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
		os.Exit(1)
	}
}

func pullIDL() error {

	os.Chdir(idlDir)

	cmd := exec.Command("git", "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
		return err
	}

	return nil
}


func NewIDLFolder(prod, srv string) *IDLFolder {
	idl := &IDLFolder{
		ProdMap: map[string]*ProdIDLFolder{},
		fileMap: map[string]struct{}{},
		conf: new(Config),
		prod:prod,
		srv:srv,
		Files: map[string]*IDLFile{},
	}
	return idl
}