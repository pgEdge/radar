%global debug_package %{nil}
%global sname pgedge-radar

Name:           %{sname}
Version:        %{radar_version}
Release:        %{radar_buildnum}%{?dist}
Summary:        Agentless, zero-dependency diagnostic data collection tool for PostgreSQL and system metrics

License:        PostgreSQL License
URL:            https://github.com/pgEdge/radar

Source0:        radar-linux-%{arch}
Source1:        LICENCE.md
Source2:        README.md

%description
radar is an agentless, zero-dependency diagnostic data collection tool for
PostgreSQL and system metrics. It collects comprehensive snapshots of system
and PostgreSQL metrics for troubleshooting and analysis, streaming all data
directly into a timestamped ZIP file without requiring agents or background
processes.

%prep
# Stage the radar binary (Source0) into the build dir alongside the docs so the
# SBOM below scans the actual Go binary (syft's go-module-binary cataloger),
# not just LICENCE/README. The install section still installs the binary from
# Source0 directly, so this copy only feeds the syft scan.
cp %{SOURCE0} %{_builddir}/radar
cp %{SOURCE1} %{_builddir}/
cp %{SOURCE2} %{_builddir}/

%build
syft dir:%{_builddir} -o cyclonedx-json > %{_builddir}/%{sname}-sbom.json || exit 1

KEY_ID=$(gpg --list-secret-keys --with-colons | awk -F: '/^sec/{print $5}' | head -n 1); export KEY_ID
gpg --armor --detach-sign --output %{_builddir}/%{sname}-sbom.json.asc %{_builddir}/%{sname}-sbom.json || exit 1

%install
install -D -m 0755 %{SOURCE0} %{buildroot}/usr/bin/radar
mkdir -p %{buildroot}%{_datadir}/%{sname}
install -p -m 0644 %{_builddir}/%{sname}-sbom.json %{buildroot}%{_datadir}/%{sname}/%{sname}-sbom.json
install -p -m 0644 %{_builddir}/%{sname}-sbom.json.asc %{buildroot}%{_datadir}/%{sname}/%{sname}-sbom.json.asc
mkdir -p %{buildroot}%{_docdir}/%{sname}
install -p -m 0644 %{_builddir}/README.md %{buildroot}%{_docdir}/%{sname}/README.md

%files
%license LICENCE.md
%{_docdir}/%{sname}/README.md
%{_bindir}/radar
%{_datadir}/%{sname}/%{sname}-sbom.json
%{_datadir}/%{sname}/%{sname}-sbom.json.asc

%changelog
* Thu Apr 16 2026 pgEdge Build Team <support@pgedge.com> - 0.4.1
- Initial RPM package of pgedge-radar
