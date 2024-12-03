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

function transform (log_rec, tz_shift_min)
    local out = new_out_rec()
    --local out = transform_default(log_rec, tz_shift_min)
    logger.warn("processing record", {foo="bar"})
    set_out_prop(out, "Type", "kontext-mod")
    set_out_prop(out, "IPAddress", "192.168.1.3")
    set_out_prop(out, "Action", string.format("*** %s::modified: %s", log_rec.Action, log_rec.x))
    return out
end