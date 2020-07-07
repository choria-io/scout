ECHO OFF

"\Program Files (x86)\WiX Toolset v3.11\bin\candle.exe" scout.wxs
"\Program Files (x86)\WiX Toolset v3.11\bin\light.exe" -ext WixUIExtension scout.wixobj -o {{cpkg_name}}-{{cpkg_version}}.msi
