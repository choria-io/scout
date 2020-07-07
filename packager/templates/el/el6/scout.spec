%define debug_package %{nil}
%define pkgname {{cpkg_name}}
%define version {{cpkg_version}}
%define bindir {{cpkg_bindir}}
%define etcdir {{cpkg_etcdir}}
%define release {{cpkg_release}}
%define dist {{cpkg_dist}}
%define manage_conf {{cpkg_manage_conf}}
%define binary {{cpkg_binary}}
%define tarball {{cpkg_tarball}}
%define contact {{cpkg_contact}}
%define pkggroup {{cpkg_rpm_group}}

Name: %{pkgname}
Version: %{version}
Release: %{release}.%{dist}
Summary: Choria Scout Monitoring Framework
License: Apache-2.0
URL: https://choria.io
Group: %{pkggroup}
Source0: %{tarball}
Packager: %{contact}
BuildRoot: %{_tmppath}/%{pkgname}-%{version}-%{release}-root-%(%{__id_u} -n)

%description
The Choria Scout Monitoring Framework

Please visit https://choria.io for more information

%prep
%setup -q

%build

%install
rm -rf %{buildroot}
%{__install} -d -m0755  %{buildroot}/etc/sysconfig
%{__install} -d -m0755  %{buildroot}/etc/logrotate.d
%{__install} -d -m0755  %{buildroot}/etc/rc.d/init.d
%{__install} -d -m0755  %{buildroot}%{bindir}
%{__install} -d -m0755  %{buildroot}%{etcdir}
%{__install} -d -m0755  %{buildroot}/var/log
%{__install} -m0755 dist/agent.init %{buildroot}/etc/rc.d/init.d/%{pkgname}-agent
%{__install} -m0644 dist/agent.sysconfig %{buildroot}/etc/sysconfig/%{pkgname}-agent
%{__install} -m0755 dist/agent-logrotate %{buildroot}/etc/logrotate.d/%{pkgname}
%if 0%{?manage_conf} > 0
%{__install} -m0640 dist/scout.conf %{buildroot}%{etcdir}/scout.conf
%endif
%{__install} -m0755 %{binary} %{buildroot}%{bindir}/%{pkgname}

%clean
rm -rf %{buildroot}

%post
/sbin/chkconfig --add %{pkgname}-agent || :

%postun
if [ "$1" -ge 1 ]; then
  /sbin/service %{pkgname}-agent condrestart &>/dev/null || :
fi

%preun
if [ "$1" = 0 ] ; then
  /sbin/service %{pkgname}-agent stop > /dev/null 2>&1
  /sbin/chkconfig --del %{pkgname}-agent || :
fi

%files
%if 0%{?manage_conf} > 0
%config(noreplace)%{etcdir}/scout.conf
%endif
%{bindir}/%{pkgname}
/etc/logrotate.d/%{pkgname}
/etc/rc.d/init.d/%{pkgname}-agent
%config(noreplace)/etc/sysconfig/%{pkgname}-agent

%changelog
* Tue Jul 07 2020 R.I.Pienaar <rip@devco.net>
- Initial Release
