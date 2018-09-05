
pipeline {
    agent any,
    stages {
        stage("checkout"){
            steps{
                scm checkout
                sh 'ls -al'
            }
        }
    }
}