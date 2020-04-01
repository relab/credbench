/**
 * Based on: https://github.com/livepeer/protocol/blob/master/scripts/parseArtifacts.js
*/
const fs = require("fs")
const path = require("path")
const mkdirp = require("mkdirp")
const ARTIFACT_DIR = path.resolve(__dirname, "../build/contracts")
const ABI_DIR = path.resolve(__dirname, "../build/abi")
const BIN_DIR = path.resolve(__dirname, "../build/bin")
mkdirp.sync(ABI_DIR)
mkdirp.sync(BIN_DIR)

const parseArtifact = (inFile, outFile, type = "abi") => {
    fs.readFile(inFile, (err, data) => {
        if (err) {
            console.error("Failed to read " + inFile + ": " + err)
        } else {
            const json = JSON.parse(data)
            let jsonData
            if (type == "bytecode") {
                jsonData = json.bytecode
            } else {
                jsonData = JSON.stringify(json.abi)
            }
            fs.writeFile(outFile, jsonData, err => {
                if (err) {
                    console.error("Failed to write " + outFile + ": " + err)
                }
            });
        }
    })
}

fs.readdir(ARTIFACT_DIR, (err, files) => {
    if (err) {
        console.error("Failed to read " + ARTIFACT_DIR + ": " + err)
    } else {
        files.forEach(filename => {
            const artifactFile = path.join(ARTIFACT_DIR, filename)
            const abiFile = path.join(ABI_DIR, path.basename(filename, ".json") + ".abi")
            const binFile = path.join(BIN_DIR, path.basename(filename, ".json") + ".bin")

            parseArtifact(artifactFile, abiFile, "abi")
            parseArtifact(artifactFile, binFile, "bytecode")
        })
    }
})