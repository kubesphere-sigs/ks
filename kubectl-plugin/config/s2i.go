package config

import (
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
)

type s2iOption struct {
	DryRun bool
}

func newS2ICmd() (cmd *cobra.Command) {
	opt := &s2iOption{}

	cmd = &cobra.Command{
		Use:   "s2i",
		Short: "Load S2iBuilderTemplates",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opt.DryRun, "dry-run", "", false, "Simulate a install of s2i builder templates")
	return
}

func (o *s2iOption) runE(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		for name := range templates {
			tplPath := path.Join(os.TempDir(), fmt.Sprintf("%s.yaml", name))
			_ = os.RemoveAll(tplPath)
		}
	}()

	for name, tpl := range templates {
		if o.DryRun {
			cmd.Println("install template", name, "the template is:\n", tpl)
		} else {
			tplPath := path.Join(os.TempDir(), fmt.Sprintf("%s.yaml", name))

			if err = ioutil.WriteFile(tplPath, []byte(tpl), 0644); err != nil {
				return
			}
			if err = common.ExecCommand("kubectl", "apply", "-f", tplPath); err != nil {
				return
			}
		}
	}
	return
}

var templates = map[string]string{
	`binary`: `apiVersion: devops.kubesphere.io/v1alpha1
kind: S2iBuilderTemplate
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
    builder-type.kubesphere.io/b2i: "b2i"
    binary-type.kubesphere.io: "binary"
  annotations:
    descriptionEN: "This is a builder template for binary build"
    descriptionCN: "二进制文件的构建器模版"
    devops.kubesphere.io/s2i-template-url: https://github.com/kubesphere/s2i-binary-container
  name: binary
spec:
  containerInfo:
    - builderImage: kubesphere/s2i-binary:v2.1.0
  environment:
    - key: ARGS
      type: string
      description: "Arguments to use when calling binary,"
      required: false
      defaultValue: ""
  codeFramework: binary
  defaultBaseImage: kubesphere/s2i-binary:v2.1.0
  version: 0.0.1
  description: "This is a builder template for binary build"
  iconPath: assets/binary.png
`,
	`java`: `apiVersion: devops.kubesphere.io/v1alpha1
kind: S2iBuilderTemplate
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
    builder-type.kubesphere.io/s2i: "s2i"
    builder-type.kubesphere.io/b2i: "b2i"
    binary-type.kubesphere.io: "jar"
  annotations:
    "helm.sh/hook": pre-install
    descriptionCN: "Java 应用的构建器模版。通过该模版可构建出直接运行的应用镜像。"
    descriptionEN: "This is a builder template for Java builds whose result can be run directly without any further application server.It's suited ideally for microservices with a flat classpath (including \"far jars\")."
    devops.kubesphere.io/s2i-template-url: https://github.com/kubesphere/s2i-java-container/blob/master/java/images
  name: java
spec:
  containerInfo:
     - builderImage: kubesphere/java-8-centos7:v2.1.0
       runtimeImage: kubesphere/java-8-runtime:v2.1.0
       runtimeArtifacts:
         - source: "/deployments"
       buildVolumes: ["s2i_java_cache:/tmp/artifacts"]
     - builderImage: kubesphere/java-11-centos7:v2.1.0
       runtimeImage: kubesphere/java-11-runtime:v2.1.0
       runtimeArtifacts:
         - source: "/deployments"
       buildVolumes: ["s2i_java_cache:/tmp/artifacts"]
  environment:
    - key: MAVEN_ARGS
      type: string
      description: "Arguments to use when calling Maven, replacing the default package hawt-app:build -DskipTests -e. Please be sure to run the hawt-app:build goal (when not already bound to the package execution phase), otherwise the startup scripts won't work."
      required: false
      defaultValue: ""
    - key: MAVEN_ARGS_APPEND
      type: string
      description: "Additional Maven arguments, useful for temporary adding arguments like -X or -am -pl ."
      required: false
      defaultValue: ""
    - key: ARTIFACT_DIR
      type: string
      description: "Path to target/ where the jar files are created for multi module builds. These are added to ${MAVEN_ARGS}"
      required: false
      defaultValue: ""
    - key: ARTIFACT_COPY_ARGS
      type: string
      description: "Arguments to use when copying artifacts from the output dir to the application dir. Useful to specify which artifacts will be part of the image. It defaults to -r hawt-app/* when a hawt-app dir is found on the build directory, otherwise jar files only will be included (*.jar)."
      required: false
      defaultValue: ""
    - key: MAVEN_CLEAR_REPO
      type: boolean
      description: "If set then the Maven repository is removed after the artifact is built. This is useful for keeping the created application image small, but prevents incremental builds. The default is false"
      required: false
      defaultValue: ""
    - key: JAVA_APP_DIR
      type: string
      description: "the directory where the application resides. All paths in your application are relative to this directory. By default it is the same directory where this startup script resides."
      required: false
      defaultValue: ""
    - key: JAVA_LIB_DIR
      type: string
      description: "directory holding the Java jar files as well an optional classpath file which holds the classpath. Either as a single line classpath (colon separated) or with jar files listed line-by-line. If not set JAVA_LIB_DIR is the same as JAVA_APP_DIR."
      required: false
      defaultValue: ""
    - key: JAVA_OPTIONS
      type: string
      description: "options to add when calling java"
      required: false
      defaultValue: ""
    - key: JAVA_MAJOR_VERSION
      type: string
      description: "a number >= 7. If the version is set then only options suitable for this version are used. When set to 7 options known only to Java > 8 will be removed. For versions >= 10 no explicit memory limit is calculated since Java >= 10 has support for container limits."
      required: false
      defaultValue: ""
    - key: JAVA_MAX_MEM_RATIO
      type: string
      description: "is used when no -Xmx option is given in JAVA_OPTIONS. This is used to calculate a default maximal Heap Memory based on a containers restriction. If used in a Docker container without any memory constraints for the container then this option has no effect. If there is a memory constraint then -Xmx is set to a ratio of the container available memory as set here. The default is 25 when the maximum amount of memory available to the container is below 300M, 50 otherwise, which means in that case that 50% of the available memory is used as an upper boundary. You can skip this mechanism by setting this value to 0 in which case no -Xmx option is added."
      required: false
      defaultValue: ""
    - key: JAVA_INIT_MEM_RATIO
      type: string
      description: "is used when no -Xms option is given in JAVA_OPTIONS. This is used to calculate a default initial Heap Memory based on a containers restriction. If used in a Docker container without any memory constraints for the container then this option has no effect. If there is a memory constraint then -Xms is set to a ratio of the container available memory as set here. By default this value is not set."
      required: false
      defaultValue: ""
    - key: JAVA_MAX_CORE
      type: string
      description: "restrict manually the number of cores available which is used for calculating certain defaults like the number of garbage collector threads. If set to 0 no base JVM tuning based on the number of cores is performed."
      required: false
      defaultValue: ""
    - key: JAVA_DIAGNOSTICS
      type: string
      description: "set this to get some diagnostics information to standard out when things are happening"
      required: false
      defaultValue: ""
    - key: JAVA_MAIN_CLASS
      type: string
      description: "main class to use as argument for java. When this environment variable is given, all jar files in $JAVA_APP_DIR are added to the classpath as well as $JAVA_LIB_DIR."
      required: false
      defaultValue: ""
    - key: JAVA_APP_JAR
      type: string
      description: "A jar file with an appropriate manifest so that it can be started with java -jar if no $JAVA_MAIN_CLASS is set. In all cases this jar file is added to the classpath, too."
      required: false
      defaultValue: ""
    - key: JAVA_APP_NAME
      type: string
      description: "Name to use for the process"
      required: false
      defaultValue: ""
    - key: JAVA_CLASSPATH
      type: string
      description: "the classpath to use. If not given, the startup script checks for a file ${JAVA_APP_DIR}/classpath and use its content literally as classpath. If this file doesn't exists all jars in the app dir are added (classes:${JAVA_APP_DIR}/*)."
      required: false
      defaultValue: ""
    - key: JAVA_DEBUG
      type: string
      description: "If set remote debugging will be switched on"
      required: false
      defaultValue: ""
    - key: JAVA_DEBUG_SUSPEND
      type: string
      description: "If set enables suspend mode in remote debugging"
      required: false
      defaultValue: ""
    - key: JAVA_DEBUG_PORT
      type: string
      description: "Port used for remote debugging. Default: 5005"
      required: false
      defaultValue: ""
    - key: HTTP_PROXY
      type: string
      description: "The URL of the proxy server that translates into the http.proxyHost and http.proxyPort system properties."
      required: false
      defaultValue: ""
    - key: HTTPS_PROXY
      type: string
      description: "The URL of the proxy server that translates into the https.proxyHost and https.proxyPort system properties."
      required: false
      defaultValue: ""
    - key: NO_PROXY
      type: string
      description: "The list of hosts that should be reached directly, bypassing the proxy, that translates into the http.nonProxyHosts system property."
      required: false
      defaultValue: ""
  codeFramework: java
  defaultBaseImage: {{ java8centos7_repo }}:v2.1.0
  version: 0.0.1
  description: "This is a builder template for Java builds whose result can be run directly without any further application server.It's suited ideally for microservices with a flat classpath (including \"far jars\")"
  iconPath: assets/java.png
`,
	`nodejs`: `apiVersion: devops.kubesphere.io/v1alpha1
kind: S2iBuilderTemplate
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
    builder-type.kubesphere.io/s2i: "s2i"
  annotations:
    descriptionEN: "Node.js available as container is a base platform for building and running various Node.js applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices."
    descriptionCN: "Nodejs 应用的构建器模版。Node.js 是基于 Chrome 的 JavaScript 运行时构建的平台，可轻松构建快速，可扩展的网络应用程序。"
    devops.kubesphere.io/s2i-template-url: https://github.com/kubesphere/s2i-nodejs-container
    "helm.sh/hook": pre-install
  name: nodejs
spec:
  containerInfo:
     - builderImage: kubesphere/nodejs-8-centos7:v2.1.0
     - builderImage: kubesphere/nodejs-6-centos7:v2.1.0
     - builderImage: kubesphere/nodejs-4-centos7:v2.1.0
  environment:
    - key: NODE_ENV
      type: string
      description: "NodeJS runtime mode (default: \"production\")"
      required: false
      defaultValue: ""
    - key: DEV_MODE
      type: boolean
      description: "When set to \"true\", nodemon will be used to automatically reload the server while you work (default: \"false\"). Setting DEV_MODE to \"true\" will change the NODE_ENV default to \"development\" (if not explicitly set)."
      required: false
      defaultValue: ""
    - key: NPM_RUN
      type: string
      description: "Select an alternate / custom runtime mode, defined in your package.json file's scripts section (default: npm run \"start\"). These user-defined run-scripts are unavailable while DEV_MODE is in use."
      required: false
      defaultValue: ""
    - key: HTTP_PROXY
      type: string
      description: "Use an npm proxy during assembly"
      required: false
      defaultValue: ""
    - key: HTTPS_PROXY
      type: string
      description: "Use an npm proxy during assembly"
      required: false
      defaultValue: ""
    - key: NPM_MIRROR
      type: string
      description: "Use a custom NPM registry mirror to download packages during the build process"
      required: false
      defaultValue: ""
    - key: YARN_ENABLED
      type: string
      description: "Set this variable to a non-empty value to use \"yarn install\" get dependencies"
      required: false
      defaultValue: ""
    - key: YARN_ARGS
      type: string
      description: "\"yarn install\" args"
      required: false
      defaultValue: ""
  codeFramework: nodejs
  defaultBaseImage: kubesphere/nodejs-8-centos7:v2.1.0
  version: 0.0.1
  description: "Node.js available as container is a base platform for building and running various Node.js applications and frameworks. Node.js is a platform built on Chrome's JavaScript runtime for easily building fast, scalable network applications. Node.js uses an event-driven, non-blocking I/O model that makes it lightweight and efficient, perfect for data-intensive real-time applications that run across distributed devices."
  iconPath: assets/nodejs.png
`,
	`python`: `apiVersion: devops.kubesphere.io/v1alpha1
kind: S2iBuilderTemplate
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
    builder-type.kubesphere.io/s2i: "s2i"
  name: python
  annotations:
    descriptionCN: "Python 应用的构建器模版。Python 是一种易于学习，功能强大的编程语言。 它具有高效的高级数据结构和简单但有效的面向对象编程方法。"
    descriptionEN: "Python available as container is a base platform for building and running various Python applications and frameworks. Python is an easy to learn, powerful programming language. It has efficient high-level data structures and a simple but effective approach to object-oriented programming. Python's elegant syntax and dynamic typing, together with its interpreted nature, make it an ideal language for scripting and rapid application development in many areas on most platforms."
    "helm.sh/hook": pre-install
    devops.kubesphere.io/s2i-template-url: https://github.com/kubesphere/s2i-python-container
spec:
  containerInfo:
    - builderImage: kubesphere/python-36-centos7:v2.1.0
    - builderImage: kubesphere/python-35-centos7:v2.1.0
    - builderImage: kubesphere/python-34-centos7:v2.1.0
    - builderImage: kubesphere/python-27-centos7:v2.1.0
  environment:
    - key: APP_SCRIPT
      type: string
      description: "Used to run the application from a script file. This should be a path to a script file (defaults to app.sh unless set to null) that will be run to start the application."
      required: false
      defaultValue: ""
    - key: APP_FILE
      description: "Used to run the application from a Python script. This should be a path to a Python file (defaults to app.py unless set to null) that will be passed to the Python interpreter to start the application."
      required: false
      defaultValue: ""
      type: string
    - key: APP_MODULE
      description: "Used to run the application with Gunicorn, as documented here. This variable specifies a WSGI callable with the pattern MODULE_NAME:VARIABLE_NAME, where MODULE_NAME is the full dotted path of a module, and VARIABLE_NAME refers to a WSGI callable inside the specified module. Gunicorn will look for a WSGI callable named application if not specified.

                      If APP_MODULE is not provided, the run script will look for a wsgi.py file in your project and use it if it exists.

                      If using setup.py for installing the application, the MODULE_NAME part can be read from there. "
      required: false
      defaultValue: ""
      type: string
    - key: APP_HOME
      description: "This variable can be used to specify a sub-directory in which the application to be run is contained. The directory pointed to by this variable needs to contain wsgi.py (for Gunicorn) or manage.py (for Django).

                      If APP_HOME is not provided, the assemble and run scripts will use the application's root directory."
      required: false
      defaultValue: ""
      type: string
    - key: APP_CONFIG
      description: "Path to a valid Python file with a Gunicorn configuration file."
      required: false
      defaultValue: ""
      type: string
    - key: DISABLE_COLLECTSTATIC
      description: "Set this variable to a non-empty value to inhibit the execution of 'manage.py collectstatic' during the build. This only affects Django projects."
      required: false
      defaultValue: ""
      type: string
    - key: DISABLE_SETUP_PY_PROCESSING
      description: "Set this to a non-empty value to skip processing of setup.py script if you use -e . in requirements.txt to trigger its processing or you don't want your application to be installed into site-packages directory."
      required: false
      defaultValue: ""
      type: string
    - key: ENABLE_PIPENV
      description: "Set this variable to use Pipenv, the higher-level Python packaging tool, to manage dependencies of the application. This should be used only if your project contains properly formated Pipfile and Pipfile.lock."
      required: false
      defaultValue: ""
      type: string
    - key: PIP_INDEX_URL
      description: "Set this variable to use a custom index URL or mirror to download required packages during build process. This only affects packages listed in requirements.txt. Pipenv ignores this variable."
      required: false
      defaultValue: ""
      type: string
    - key: UPGRADE_PIP_TO_LATEST
      description: "Set this variable to a non-empty value to have the 'pip' program and related python packages (setuptools and wheel) be upgraded to the most recent version before any Python packages are installed. If not set it will use whatever the default version is included by the platform for the Python version being used.."
      required: false
      defaultValue: ""
      type: string
    - key: WEB_CONCURRENCY
      description: "Set this to change the default setting for the number of workers. By default, this is set to the number of available cores times 2, capped at 12."
      required: false
      defaultValue: ""
      type: string
  codeFramework: python
  defaultBaseImage: kubesphere/python-36-centos7:v2.1.0
  version: 0.0.1
  description: "Python available as container is a base platform for building and running various Python applications and frameworks. Python is an easy to learn, powerful programming language. It has efficient high-level data structures and a simple but effective approach to object-oriented programming. Python's elegant syntax and dynamic typing, together with its interpreted nature, make it an ideal language for scripting and rapid application development in many areas on most platforms.

                This container image includes an npm utility, so users can use it to install JavaScript modules for their web applications. There is no guarantee for any specific npm or nodejs version, that is included in the image; those versions can be changed anytime and the nodejs itself is included just to make the npm work."
  iconPath: assets/python.png
`,
	`tomcat`: `apiVersion: devops.kubesphere.io/v1alpha1
kind: S2iBuilderTemplate
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
    builder-type.kubesphere.io/s2i: "s2i"
    builder-type.kubesphere.io/b2i: "b2i"
    binary-type.kubesphere.io: "war"
  name: tomcat
  annotations:
    descriptionCN: "Tomcat 应用的构建器模版，通过该模版可构建出直接运行的应用镜像。"
    descriptionEN: "This is a builder template for Java builds whose result can be run directly with Tomcat application server."
    "helm.sh/hook": pre-install
    devops.kubesphere.io/s2i-template-url: https://github.com/kubesphere/s2i-java-container/tree/master/tomcat/images/
spec:
  containerInfo:
    - builderImage: kubesphere/tomcat85-java11-centos7:v2.1.0
      runtimeImage: kubesphere/tomcat85-java11-runtime:v2.1.0
      runtimeArtifacts:
        - source: "/deployments"
      buildVolumes: ["s2i_java_cache:/tmp/artifacts"]
    - builderImage: kubesphere/tomcat85-java8-centos7:v2.1.0
      runtimeImage: kubesphere/tomcat85-java8-runtime:v2.1.0
      runtimeArtifacts:
        - source: "/deployments"
      buildVolumes: ["s2i_java_cache:/tmp/artifacts"]
  environment:
    - key: MAVEN_ARGS
      type: string
      description: "Arguments to use when calling Maven, replacing the default package hawt-app:build -DskipTests -e. Please be sure to run the hawt-app:build goal (when not already bound to the package execution phase), otherwise the startup scripts won't work."
      required: false
      defaultValue: ""
    - key: MAVEN_ARGS_APPEND
      type: string
      description: "Additional Maven arguments, useful for temporary adding arguments like -X or -am -pl ."
      required: false
      defaultValue: ""
    - key: ARTIFACT_DIR
      type: string
      description: "Path to target/ where the jar files are created for multi module builds. These are added to ${MAVEN_ARGS}"
      required: false
      defaultValue: ""
    - key: ARTIFACT_COPY_ARGS
      type: string
      description: "Arguments to use when copying artifacts from the output dir to the application dir. Useful to specify which artifacts will be part of the image. It defaults to -r hawt-app/* when a hawt-app dir is found on the build directory, otherwise jar files only will be included (*.jar)."
      required: false
      defaultValue: ""
    - key: MAVEN_CLEAR_REPO
      type: boolean
      description: "If set then the Maven repository is removed after the artifact is built. This is useful for keeping the created application image small, but prevents incremental builds. The default is false"
      required: false
      defaultValue: ""
    - key: JAVA_APP_DIR
      type: string
      description: "the directory where the application resides. All paths in your application are relative to this directory. By default it is the same directory where this startup script resides."
      required: false
      defaultValue: ""
    - key: JAVA_LIB_DIR
      type: string
      description: "directory holding the Java jar files as well an optional classpath file which holds the classpath. Either as a single line classpath (colon separated) or with jar files listed line-by-line. If not set JAVA_LIB_DIR is the same as JAVA_APP_DIR."
      required: false
      defaultValue: ""
    - key: JAVA_OPTIONS
      type: string
      description: "options to add when calling java"
      required: false
      defaultValue: ""
    - key: JAVA_MAJOR_VERSION
      type: string
      description: "a number >= 7. If the version is set then only options suitable for this version are used. When set to 7 options known only to Java > 8 will be removed. For versions >= 10 no explicit memory limit is calculated since Java >= 10 has support for container limits."
      required: false
      defaultValue: ""
    - key: JAVA_MAX_MEM_RATIO
      type: string
      description: "is used when no -Xmx option is given in JAVA_OPTIONS. This is used to calculate a default maximal Heap Memory based on a containers restriction. If used in a Docker container without any memory constraints for the container then this option has no effect. If there is a memory constraint then -Xmx is set to a ratio of the container available memory as set here. The default is 25 when the maximum amount of memory available to the container is below 300M, 50 otherwise, which means in that case that 50% of the available memory is used as an upper boundary. You can skip this mechanism by setting this value to 0 in which case no -Xmx option is added."
      required: false
      defaultValue: ""
    - key: JAVA_INIT_MEM_RATIO
      type: string
      description: "is used when no -Xms option is given in JAVA_OPTIONS. This is used to calculate a default initial Heap Memory based on a containers restriction. If used in a Docker container without any memory constraints for the container then this option has no effect. If there is a memory constraint then -Xms is set to a ratio of the container available memory as set here. By default this value is not set."
      required: false
      defaultValue: ""
    - key: JAVA_MAX_CORE
      type: string
      description: "restrict manually the number of cores available which is used for calculating certain defaults like the number of garbage collector threads. If set to 0 no base JVM tuning based on the number of cores is performed."
      required: false
      defaultValue: ""
    - key: JAVA_DIAGNOSTICS
      type: string
      description: "set this to get some diagnostics information to standard out when things are happening"
      required: false
      defaultValue: ""
    - key: JAVA_MAIN_CLASS
      type: string
      description: "main class to use as argument for java. When this environment variable is given, all jar files in $JAVA_APP_DIR are added to the classpath as well as $JAVA_LIB_DIR."
      required: false
      defaultValue: ""
    - key: JAVA_APP_JAR
      type: string
      description: "A jar file with an appropriate manifest so that it can be started with java -jar if no $JAVA_MAIN_CLASS is set. In all cases this jar file is added to the classpath, too."
      required: false
      defaultValue: ""
    - key: JAVA_APP_NAME
      type: string
      description: "Name to use for the process"
      required: false
      defaultValue: ""
    - key: JAVA_CLASSPATH
      type: string
      description: "the classpath to use. If not given, the startup script checks for a file ${JAVA_APP_DIR}/classpath and use its content literally as classpath. If this file doesn't exists all jars in the app dir are added (classes:${JAVA_APP_DIR}/*)."
      required: false
      defaultValue: ""
    - key: JAVA_DEBUG
      type: string
      description: "If set remote debugging will be switched on"
      required: false
      defaultValue: ""
    - key: JAVA_DEBUG_SUSPEND
      type: string
      description: "If set enables suspend mode in remote debugging"
      required: false
      defaultValue: ""
    - key: JAVA_DEBUG_PORT
      type: string
      description: "Port used for remote debugging. Default: 5005"
      required: false
      defaultValue: ""
    - key: HTTP_PROXY
      type: string
      description: "The URL of the proxy server that translates into the http.proxyHost and http.proxyPort system properties."
      required: false
      defaultValue: ""
    - key: HTTPS_PROXY
      type: string
      description: "The URL of the proxy server that translates into the https.proxyHost and https.proxyPort system properties."
      required: false
      defaultValue: ""
    - key: NO_PROXY
      type: string
      description: "The list of hosts that should be reached directly, bypassing the proxy, that translates into the http.nonProxyHosts system property."
      required: false
      defaultValue: ""
  codeFramework: java
  defaultBaseImage: kubesphere/tomcat85-java8-centos7:v2.1.0
  version: 0.0.1
  description: "This is a builder template for Java builds whose result can be run directly with Tomcat application server."
  iconPath: assets/tomcat.png
`,
}
