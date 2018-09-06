properties([disableConcurrentBuilds()])

//12
node("dind") {
    container('dind') {
        compat.test_build()
    }
}
