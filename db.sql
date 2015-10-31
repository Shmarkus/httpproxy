CREATE database httpproxy;
CREATE SCHEMA proxy;
CREATE TABLE proxy.mapping (needle VARCHAR(20), mock VARCHAR(1000));
CREATE USER http password 'proxy';
GRANT ALL ON SCHEMA proxy TO http;
GRANT ALL ON ALL TABLES IN SCHEMA proxy TO http;