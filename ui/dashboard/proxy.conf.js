const PROXY_CONFIG = {};
const baseUrl = process.env.PREFIX_API_BASE_URL || '/api/';
const replaceBaseUrl = process.env.REPLACE_BASE_URL || '/';
const proxyUrl = process.env.PROXY_URL || 'http://localhost:8081';

PROXY_CONFIG[baseUrl] = {
    "target": proxyUrl,
    "pathRewrite": {},
    "secure": false,
    "logLevel": "debug"
};
PROXY_CONFIG[baseUrl].pathRewrite[`^${baseUrl}`] = replaceBaseUrl;

module.exports = PROXY_CONFIG;
