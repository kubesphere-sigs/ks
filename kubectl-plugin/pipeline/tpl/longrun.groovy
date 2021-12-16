pipeline {
  agent any

  stages {
    stage('one') {
      steps {
        script {
          for (i = 0; i < 60; i++) {
            sleep 1
            echo "1"
          }
        }
      }
    }
    stage('two') {
      steps {
        script {
          for (i = 0; i < 60; i++) {
            sleep 1
            echo "1"
          }
        }
      }
    }
    stage('three') {
      steps {
        script {
          for (i = 0; i < 60; i++) {
            sleep 1
            echo "1"
          }
        }
      }
    }
  }
}
