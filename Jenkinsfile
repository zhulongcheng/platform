
pipeline {
    agent {docker:'dind'}
    stages {
        stage("checkout"){
            steps{
                scm checkout
                sh ''
            }
        }
    }
}