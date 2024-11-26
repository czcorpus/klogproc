--[[
Input Record Fields:

Logger (string)
Level (string)
Date (string)
Message (string)
Exception (table (struct))
    ID (string)
    Type (string)
    Stack (table (array)) - Contents: string
UserID (number (integer))
ProcTime (number (float))
Action (string)
IsIndirectCall (boolean)
Request (table (struct))
    HTTPForwardedFor (string)
    HTTPUserAgent (string)
    HTTPRemoteAddr (string)
    RemoteAddr (string)
Args (table (map)) - Contents: key=string, value=interface {}
Error (table (struct))
    Name (string)
    Anchor (string)
]]--

dmp = require('dump')

function transform (logRec, tzShiftMin)
    --local out = outputRecord.new()
    print(dmp(default_transformer))
    local out = default_transformer:transform(logRec, tzShiftMin)
    --print("Lua func. called...")
    set_out(out, "Type", "kontext")
    set_out(out, "Action", string.format("*** %s::modified: %s", logRec.Action, logRec.x))
    return out
end