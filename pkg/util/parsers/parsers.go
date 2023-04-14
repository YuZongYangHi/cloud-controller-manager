package parsers

import (
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v2"
	"io/fs"
	"os"
	"strconv"
)

const (
	YAML = "yaml"
	JSON = "json"
)

func ParserConfigurationByFile(format, in string, out interface{}) error {
	data, err := fs.ReadFile(os.DirFS("."), in)

	if err != nil {
		return err
	}

	switch format {
	case YAML:
		return yaml.Unmarshal(data, out)
	case JSON:
		return json.Unmarshal(data, out)
	default:
		return errors.New("invalid file format")
	}
}

func ParserInt64(in string) (int64, error) {
	return strconv.ParseInt(in, 10, 64)
}

func JsonLoads(data []byte, obj interface{}) error {

	var err error

	if err = json.Unmarshal(data, obj); err != nil {
		return err
	}
	return nil
}

func JsonDumps(i interface{}) ([]byte, error) {
	var b []byte
	var err error

	if b, err = json.Marshal(i); err != nil {
		return nil, err
	}
	return b, nil
}

func JsonInterface(i interface{}, obj interface{}) error {
	var err error
	var b []byte

	if b, err = JsonDumps(i); err == nil {
		return JsonLoads(b, obj)
	}
	return err
}
