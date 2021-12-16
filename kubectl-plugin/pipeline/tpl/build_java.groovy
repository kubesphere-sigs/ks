pipeline {
  agent {
    node {
      label 'maven'
    }
  }
  stages {
    stage('Clone') {
      steps {
        git(url: 'https://github.com/kubesphere-sigs/demo-java', changelog: true, poll: false)
      }
    }
    stage('Build & Test') {
      steps {
        container('maven') {
          sh 'mvn package test'
        }
      }
    }
    stage('Code Scan') {
      steps {
        withSonarQubeEnv('sonar') {
          container('maven') {
            sh '''mvn --version
mvn sonar:sonar \\
  -Dsonar.projectKey=test \\
  -Dsonar.host.url=http://139.198.9.130:30687/ \\
  -Dsonar.login=b3e146cdb76ecef5ffb12743779cd78e69a4b5c5'''
          }

        }

        waitForQualityGate 'false'
      }
    }
    stage('Build Image') {
      steps {
        container('maven') {
          withCredentials([usernamePassword(credentialsId : 'docker' ,passwordVariable : 'PASS' ,usernameVariable : 'USER' ,)]) {
            sh '''docker login -u $USER -p $PASS
cat <<EOM >Dockerfile
FROM kubesphere/java-8-centos7:v2.1.0
COPY target/demo-java-1.0.0.jar demo.jar
COPY target/lib demo-lib
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "demo.jar"]
EOM
docker build . -t surenpi/java-demo
docker push surenpi/java-demo'''
          }
        }
      }
    }
  }
}
