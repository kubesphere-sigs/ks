package tpl

import (
	// Enable go embed
	_ "embed"
)

var (
	//go:embed longrun.groovy
	longRunPipeline string
	//go:embed build_java.groovy
	buildJava string
	//go:embed build_go.groovy
	buildGo string
	//go:embed simple.groovy
	simple string
	//go:embed parameter.groovy
	parameter string
	//go:embed parallel.groovy
	parallel string
)

// GetLongRunPipeline returns the content of long run Pipeline
func GetLongRunPipeline() string {
	return longRunPipeline
}

// GetBuildJava returns the content of java building Jenkinsfile template
func GetBuildJava() string {
	return buildJava
}

// GetBuildGo returns the content of go building Jenkinsfile template
func GetBuildGo() string {
	return buildGo
}

// GetSimple returns the content of a simple Jenkinsfile template
func GetSimple() string {
	return simple
}

// GetParameter return the content of a parameter Jenkinsfile template
func GetParameter() string {
	return parameter
}

// GetParallel returns the content of a paralleled Jenkins pipeline template
func GetParallel() string {
	return parallel
}

// GetAllTemplates returns all Pipeline templates
func GetAllTemplates() []string {
	return []string{"java", "go", "simple", "parameter", "longRun", "parallel",
		"multi-branch-gitlab", "multi-branch-github", "multi-branch-git"}
}
