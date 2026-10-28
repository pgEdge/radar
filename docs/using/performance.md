# Understanding Performance

radar streams data directly into the ZIP file without buffering the data in
memory. radar runs the collectors sequentially with a small memory footprint,
and opens one connection per database for all of the queries in that database.
A complete collection typically takes seconds.
