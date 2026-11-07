"""Build radar-fixtures.json for the radar 14-cell chain check from STAGING-channel METADATA only
(repomd.xml, primary.xml.gz, InRelease, Packages.gz); never fetches a .rpm or .deb.
Adapted from the ACE harness's ace_release_fixtures.py (same schema): per detector cell, the one
published pgedge-radar 0.6.3-1 file with its URL, sha256 and size from the repository metadata.
radar v0.6.3 routes to the staging channel (empty tag suffix -> staging in radar's release.yml),
and the release channel does not carry 0.6.3. Also records, per cell, the PostgreSQL 16/17/18
server package versions in that channel's metadata (the prerequisite PEP installs unpinned).
usage: python3 -I radar_staging_fixtures.py <out_dir>"""
import gzip, hashlib, json, os, re, subprocess, sys, time, xml.etree.ElementTree as ET

OUT = sys.argv[1]
os.makedirs(OUT, exist_ok=True)
NS = {"r": "http://linux.duke.edu/metadata/repo", "c": "http://linux.duke.edu/metadata/common"}
ALLOWED = re.compile(r"(repomd\.xml|primary\.xml\.gz|InRelease|Packages\.gz)$")
CHANNEL = "staging"
NAME, VERSION = "pgedge-radar", "0.6.3-1"
PG = ("16", "17", "18")
RPM = {"el-9": "9", "el-10": "10"}
DEB = ("jammy", "noble", "resolute", "bookworm", "trixie")
ARCH = {"amd64": "x86_64", "arm64": "aarch64"}


def fetch(url):
    assert ALLOWED.search(url), url
    r = subprocess.run(["curl", "-sS", "--fail", "--proto", "=https", "--connect-timeout", "10", "--max-time", "120",
                        "-A", "pep-radar-fixtures-meta", url], capture_output=True)
    if r.returncode != 0:
        raise SystemExit("metadata fetch failed: %s (%s)" % (url, r.stderr.decode().strip()[:200]))
    return r.stdout


FETCHED_AT = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
cells, source = {}, {}
for os_token, rel in RPM.items():
    for arch, native in ARCH.items():
        base = "https://dnf.pgedge.com/%s/%s/RPMS/%s/" % (CHANNEL, rel, native)
        repomd = fetch(base + "repodata/repomd.xml")
        root = ET.fromstring(repomd)
        data = [d for d in root.findall("r:data", NS) if d.get("type") == "primary"][0]
        href = data.find("r:location", NS).get("href")
        want = data.findtext("r:checksum", namespaces=NS)
        body = fetch(base + href)
        got = hashlib.sha256(body).hexdigest()
        prim = ET.fromstring(gzip.decompress(body))
        hits, servers = [], {}
        for p in prim.findall("c:package", NS):
            v = p.find("c:version", NS)
            name, ver = p.findtext("c:name", namespaces=NS), "%s-%s" % (v.get("ver"), v.get("rel"))
            if name == NAME and ver == "%s.el%s" % (VERSION, rel):
                loc = p.find("c:location", NS).get("href")
                hits.append({"name": NAME, "version": ver, "file": loc.rsplit("/", 1)[-1],
                             "url": base + loc, "sha256": p.findtext("c:checksum", namespaces=NS),
                             "size": int(p.find("c:size", NS).get("package")), "arch": p.findtext("c:arch", namespaces=NS)})
            for pg in PG:
                if name == "pgedge-postgresql%s-server" % pg:
                    servers.setdefault(name, []).append(ver)
        cid = "pepcell.v1.rpm.%s.%s.packaging" % (os_token, arch)
        cells[cid] = hits
        source[cid] = {"repomd_revision": root.findtext("r:revision", namespaces=NS), "primary": href,
                       "primary_sha256": got, "primary_sha256_matches_repomd": got == want,
                       "server_versions": {k: sorted(v) for k, v in sorted(servers.items())}}
for cn in DEB:
    base = "https://apt.pgedge.com/%s/%s/" % (CHANNEL, cn)
    inrel = fetch(base + "dists/%s/InRelease" % cn).decode("utf-8", "replace")
    shas = {m.group(3): m.group(1) for m in re.finditer(r"^ ([0-9a-f]{64})\s+(\d+)\s+(\S+)$",
                                                          inrel.split("\nSHA256:", 1)[-1], re.M)}
    for arch in ARCH:
        rel = "main/binary-%s/Packages.gz" % arch
        body = fetch(base + "dists/%s/%s" % (cn, rel))
        got = hashlib.sha256(body).hexdigest()
        hits, servers = [], {}
        for blk in gzip.decompress(body).decode("utf-8", "replace").split("\n\n"):
            f = dict(l.split(": ", 1) for l in blk.splitlines() if ": " in l and not l.startswith(" "))
            if f.get("Package") == NAME and f.get("Version") == "%s.%s" % (VERSION, cn):
                hits.append({"name": NAME, "version": f["Version"], "file": f["Filename"].rsplit("/", 1)[-1],
                             "url": base + f["Filename"], "sha256": f["SHA256"], "size": int(f["Size"]),
                             "arch": f.get("Architecture")})
            for pg in PG:
                if f.get("Package") == "pgedge-postgresql-%s" % pg:
                    servers.setdefault(f["Package"], []).append(f.get("Version"))
        cid = "pepcell.v1.deb.%s.%s.packaging" % (cn, arch)
        cells[cid] = hits
        source[cid] = {"packages_gz_sha256": got, "packages_gz_sha256_matches_inrelease": shas.get(rel) == got,
                       "inrelease_date": (re.search(r"^Date: (.*)$", inrel, re.M) or [None, None])[1],
                       "server_versions": {k: sorted(v) for k, v in sorted(servers.items())}}

table = {"note": ("Validation fixture table, read-only: the already-published pgedge-radar 0.6.3-1 packages "
                  "in the STAGING channel, one per detector cell, with the sha256 and size from the repository "
                  "metadata (the independent source). The package jobs download these files and check them "
                  "against it; verify_chain.py expects these sha256 values at every stage. These are "
                  "existing-package validation fixtures, not release-build outputs."),
         "channel": CHANNEL, "cells": {k: [{x: h[x] for x in ("name", "version", "file", "url", "sha256", "size")}
                                           for h in v] for k, v in sorted(cells.items())}}
table["source_table_sha256"] = hashlib.sha256(json.dumps(table["cells"], sort_keys=True).encode()).hexdigest()
json.dump(table, open(os.path.join(OUT, "radar-fixtures.json"), "w"), indent=2, sort_keys=True)
json.dump({"channel": CHANNEL, "fetched_at": FETCHED_AT, "cells": cells, "metadata": source},
          open(os.path.join(OUT, "radar-fixtures-source.json"), "w"), indent=2, sort_keys=True)
bad = {k: len(v) for k, v in cells.items() if len(v) != 1}
print("cells:", len(cells), "with exactly one file:", sum(1 for v in cells.values() if len(v) == 1), "bad:", bad)
print("metadata digests consistent:", all(s.get("primary_sha256_matches_repomd", s.get("packages_gz_sha256_matches_inrelease"))
                                         for s in source.values()))
