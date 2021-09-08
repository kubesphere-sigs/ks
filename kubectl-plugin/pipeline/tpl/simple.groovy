pipeline {
  agent any

  environment {
    APP_NAME = "this is a sample app"
  }

  stages{
    stage('simple'){
      steps{
        echo "env " + env.APP_NAME
      }
    }

    stage('debug stage') {
      when {
        equals expected: true, actual: params.debug
      }
      steps {
        echo "It's joke, there is debug info here."
      }
    }

    stage('parallel'){
      parallel {
        stage('channel-1'){
          steps{
            echo "channel-1"
          }
        }
        stage('channel-2'){
          steps{
            echo "channel-2"
          }
        }
      }
    }
  }
}
