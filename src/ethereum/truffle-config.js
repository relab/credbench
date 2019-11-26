// https://github.com/chaijs/chai/pull/868
// Using Should style globally
require('chai/register-should');

module.exports = {
    networks: {
        development: { // local test net
            host: "127.0.0.1",
            port: 8545,
            network_id: "*" // eslint-disable-line camelcase
        },
        ganache: { // ganache-cli
            host: "127.0.0.1",
            port: 7545,
            network_id: "5777" // eslint-disable-line camelcase
        },
        develop: { // truffle development
            host: "127.0.0.1",
            port: 8545,
            network_id: "*", // eslint-disable-line camelcase
            accounts: 5,
            defaultEtherBalance: 50
        },
    },

    mocha: {
        // timeout: 100000,
        useColors: true
    },

    compilers: {
        solc: {
            version: '0.5.13'
            // settings: {
            //  optimizer: {
            //    enabled: false,
            //    runs: 200
            //  },
            //  evmVersion: "byzantium"
            // }
        }
    }
};
