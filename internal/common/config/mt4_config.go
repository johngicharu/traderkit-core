package config

var MT4Config = `; common settings
Profile=
; MarketWatch=
Login={{login}}
Password={{password}}
Server={{server}}
; AutoConfiguration=
; DataServer=
EnabledDDE=false
EnabledNews=false
; MQL5Login=
; MQL5Password=
; experts settings
ExpertsEnabled=true
ExpertsDllImport=true
ExpertsExpImport=false
ExpertsTrades=true
; open chart and run expert and/or script
Symbol=TKCORE
Period=H4
Expert=TraderkitCore
ExpertParameters=TraderkitCore.set
; do not configure any symbols due to prefixes and suffixes brokers have
; Template=Default.tpl
; Script=TraderkitCoreStartup
; ScriptParameters==per_conv.set`
