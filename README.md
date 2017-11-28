#fpc-exporter

Quick and dirty prometheus exporter for checking live http endpoints.

The tool will try to GET each listed page and up to 10 links that are referenced on it. Load time and optionally a failure counter metrics would be exposed.

fpc_load_time and fpc_load_failures will be labelled:
- page - the url that was GET'd
- parent - the parent url
- statusCode - status code as string. 0 as a possible value for timeout.

Configuration is heavily use-case (my use-case :-)) oriented. The tool was supposed to query server by IPs with Host set as header.
