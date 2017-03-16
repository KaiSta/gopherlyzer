package oracler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
    "strconv"
)

type OracleAPI struct {
	FilePath   string
	Src        string
	OraclePath string
	RootPath   string
}
type PeersMode struct {
	Mode        string  `json:"mode"`
	PeerEntries Peers `json:"peers"`
}
type DefinitionMode struct {
    Mode string `json:"mode"`
    Definition Definition `json:"definition"`
}

type Definition struct {
    ObjPos string   `json:"objpos"`
    Desc    string  `json:"desc"`
}
type Peers struct {
	Pos      string   `json:"pos"`
	Type     string   `json:"type"`
	Allocs   []string `json:"allocs"`
    Sends    []string `json:"sends"`
	Receives []string `json:"receives"`
}

func New(oraclepath, filepath, rootpath string) *OracleAPI {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	return &OracleAPI{filepath, string(data), oraclepath, rootpath}
}

func (o *OracleAPI) GetPeers(format string, charpos int) PeersMode {
	cmd := exec.Command(o.OraclePath,
		fmt.Sprintf("-%v", format), fmt.Sprintf("-scope=%s", o.RootPath),
		"peers",
		fmt.Sprintf("%v:#%v", o.FilePath, charpos))
	cmd.Stderr = os.Stdout

	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	r := strings.NewReader(string(out))

	var peers PeersMode
	if err := json.NewDecoder(r).Decode(&peers); err != nil {
		panic(err)
	}
   
	return peers
}

func (o *OracleAPI) LineNColumn(text string) (line, column int64) {
    split := strings.Split(text, ":")
    l, err := strconv.ParseInt(split[1], 10, 32)
    if err != nil {
        panic(err)
    }
    c, err := strconv.ParseInt(split[2], 10, 32)
    if err != nil {
        panic(err)
    }
    return l, c
}

func (o *OracleAPI) GetOffset(line, column int64) int64 {
    currLine := 0
    currCol := 0
    for i, c := range o.Src {
        if int64(currLine) == line-1 && int64(currCol) == column {
            return int64(i)
        }
        
        if c =='\n' {
            currLine++
            currCol = 0
        } else {
            currCol++
        }   
    }
    return -1
}

func (o *OracleAPI) FindDefinition(line int64) int64 {
    currLine := 0
    currCol := 0
    for i, c := range o.Src {
        if int64(currLine) == line-1 {
            for j := i; j < len(o.Src); j++ {
                if o.Src[j] == '=' || o.Src[j] == ':' {
                    return int64(j)
                }
            }
            return -1
        }
        
        if c =='\n' {
            currLine++
            currCol = 0
        } else {
            currCol++
        }   
    }
    return -1
}

func (o *OracleAPI) FindVariableName(line int64) string {

	i := o.GetOffset(line, 0)
	j := o.FindDefinition(line)

	cmd := exec.Command(o.OraclePath,
		"-json", fmt.Sprintf("-scope=%v", o.RootPath),
		"definition",
		fmt.Sprintf("%v:#%v,#%v", o.FilePath, i, j))
	cmd.Stderr = os.Stdout
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	r := strings.NewReader(string(out))

	var def DefinitionMode
	if err := json.NewDecoder(r).Decode(&def); err != nil {
		panic(err)
	}
    
	return def.Definition.Desc
}

func (o *OracleAPI) SimplifiedVarDesc(line string) string {
    
    if strings.Contains(line, ".") {
       s1 := strings.Split(line, ".")
        s2 := strings.Split(s1[1], " ")
        return s2[0] 
    } else {
        s1 := strings.Split(line, " ")
        return s1[1]
    }
    
}