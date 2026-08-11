local UI_PREFIX = "kwords-ui"

function preprocess (log_rec, buffer)
    return {log_rec}
end

function transform (input_rec)
    local out = transform_default(input_rec)
    local client = input_rec.Headers["x-client"] or ""
    if #client > 0 then
        set_out_prop(out, "ClientFlag", client)
    end
    if client:sub(1, #UI_PREFIX) == UI_PREFIX then
        set_out_prop(out, "IsAPI", false)
    else
        set_out_prop(out, "IsAPI", true)
    end
    if is_after_datetime(out, "2025-06-05T13:00:00") then
        datetime_add_minutes(out, -120)
    end
    return out
end
