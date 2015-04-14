package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

type fileSystemDeploymentConfigService struct {
	manifestPath  string `inject:"manifestPath"`
	configPath    string
	fs            boshsys.FileSystem
	uuidGenerator boshuuid.Generator
	logger        boshlog.Logger
	logTag        string
}

func NewFileSystemDeploymentConfigService(fs boshsys.FileSystem, uuidGenerator boshuuid.Generator, logger boshlog.Logger) DeploymentConfigService {
	return &fileSystemDeploymentConfigService{
		fs:            fs,
		uuidGenerator: uuidGenerator,
		logger:        logger,
		logTag:        "config",
	}
}

func (s *fileSystemDeploymentConfigService) Path() string {
	return s.configPath
}

func (s *fileSystemDeploymentConfigService) Exists() bool {
	return s.fs.FileExists(s.configPath)
}

func (s *fileSystemDeploymentConfigService) SetConfigPath(path string) {
	s.configPath = path
}

func (s *fileSystemDeploymentConfigService) Load() (DeploymentFile, error) {
	if s.configPath == "" {
		panic("configPath not yet set!")
	}

	s.logger.Debug(s.logTag, "Loading deployment config: %s", s.configPath)

	deploymentFile := &DeploymentFile{}

	if s.fs.FileExists(s.configPath) {
		deploymentFileContents, err := s.fs.ReadFile(s.configPath)
		if err != nil {
			return DeploymentFile{}, bosherr.WrapErrorf(err, "Reading deployment config file '%s'", s.configPath)
		}
		s.logger.Debug(s.logTag, "Deployment File Contents %#s", deploymentFileContents)

		err = json.Unmarshal(deploymentFileContents, deploymentFile)
		if err != nil {
			return DeploymentFile{}, bosherr.WrapErrorf(err, "Unmarshalling deployment config file '%s'", s.configPath)
		}
	}

	err := s.initDefaults(deploymentFile)
	if err != nil {
		return DeploymentFile{}, bosherr.WrapErrorf(err, "Initializing deployment config defaults")
	}

	return *deploymentFile, nil
}

func (s *fileSystemDeploymentConfigService) Save(deploymentFile DeploymentFile) error {
	if s.configPath == "" {
		panic("configPath not yet set!")
	}

	s.logger.Debug(s.logTag, "Saving deployment config %#v", deploymentFile)

	jsonContent, err := json.MarshalIndent(deploymentFile, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling deployment config into JSON")
	}

	err = s.fs.WriteFile(s.configPath, jsonContent)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing deployment config file '%s'", s.configPath)
	}

	return nil
}

func (s *fileSystemDeploymentConfigService) initDefaults(deploymentFile *DeploymentFile) error {
	if deploymentFile.DirectorID == "" {
		uuid, err := s.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "Generating DirectorID")
		}
		deploymentFile.DirectorID = uuid

		err = s.Save(*deploymentFile)
		if err != nil {
			return bosherr.WrapError(err, "Saving deployment config")
		}
	}

	return nil
}
