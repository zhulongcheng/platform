
properties([disableConcurrentBuilds()])
pipeline{
    agent {node:"dind", container:"dind"}
        stage("checkout"){
            steps{
                scm checkout
            }
        }
    }
}