package llm

import (
	"encoding/binary"
	"errors"
	"io"
	"path"
	"sync"
)

type ModelFamily string

const ModelFamilyUnknown ModelFamily = "unknown"

type ModelType uint32

const (
	ModelType3B  ModelType = 26
	ModelType7B  ModelType = 32
	ModelType13B ModelType = 40
	ModelType34B ModelType = 48
	ModelType30B ModelType = 60
	ModelType65B ModelType = 80
)

func (mt ModelType) String() string {
	switch mt {
	case ModelType3B:
		return "3B"
	case ModelType7B:
		return "7B"
	case ModelType13B:
		return "13B"
	case ModelType34B:
		return "34B"
	case ModelType30B:
		return "30B"
	case ModelType65B:
		return "65B"
	default:
		return "Unknown"
	}
}

type FileType interface {
	String() string
}

type ModelFile struct {
	magic uint32
	container
	model
}

type model interface {
	ModelFamily() ModelFamily
	ModelType() ModelType
	FileType() FileType
}

type container interface {
	Name() string
	Decode(io.Reader) (model, error)
}

type containerGGML struct{}

func (c *containerGGML) Name() string {
	return "ggml"
}

func (c *containerGGML) Decode(r io.Reader) (model, error) {
	return nil, nil
}

type containerGGMF struct {
	version uint32
}

func (c *containerGGMF) Name() string {
	return "ggmf"
}

func (c *containerGGMF) Decode(r io.Reader) (model, error) {
	var version uint32
	binary.Read(r, binary.LittleEndian, &version)

	switch version {
	case 1:
	default:
		return nil, errors.New("invalid version")
	}

	c.version = version
	return nil, nil
}

type containerGGJT struct {
	version uint32
}

func (c *containerGGJT) Name() string {
	return "ggjt"
}

func (c *containerGGJT) Decode(r io.Reader) (model, error) {
	var version uint32
	binary.Read(r, binary.LittleEndian, &version)

	switch version {
	case 1, 2, 3:
	default:
		return nil, errors.New("invalid version")
	}

	c.version = version

	// different model types may have different layouts for hyperparameters
	var llama llamaModel
	binary.Read(r, binary.LittleEndian, &llama.hyperparameters)
	return &llama, nil
}

type containerLORA struct {
	version uint32
}

func (c *containerLORA) Name() string {
	return "ggla"
}

func (c *containerLORA) Decode(r io.Reader) (model, error) {
	var version uint32
	binary.Read(r, binary.LittleEndian, &version)

	switch version {
	case 1:
	default:
		return nil, errors.New("invalid version")
	}

	c.version = version
	return nil, nil
}

var (
	ggmlGPU = path.Join("llama.cpp", "ggml", "build", "gpu", "bin")
	ggmlCPU = path.Join("llama.cpp", "ggml", "build", "cpu", "bin")
)

var (
	ggmlInit       sync.Once
	ggmlRunnerPath string
)

func initGGML() {
	ggmlInit.Do(func() {
		ggmlRunnerPath = chooseRunner(ggmlGPU, ggmlCPU)
	})
}

func ggmlRunner() ModelRunner {
	initGGML()
	return ModelRunner{Path: ggmlRunnerPath}
}
