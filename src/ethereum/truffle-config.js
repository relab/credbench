// https://github.com/chaijs/chai/pull/868
// Using Should style globally
require('chai/register-should');

module.exports = {
    networks: {
        development: {
            host: '127.0.0.1',
            port: 8545,
            network_id: '42' // eslint-disable-line camelcase
        },
        ganache: {
            host: '127.0.0.1',
            port: 7545,
            network_id: '5777' // eslint-disable-line camelcase
        }
    },

    mocha: {
    // timeout: 100000,
        useColors: true
    },

    compilers: {
        solc: {
            version: '0.5.7'
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
