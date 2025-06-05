function preprocess (log_rec, buffer)
    return {log_rec}
end

function transform (input_rec)
    local out = transform_default(input_rec)
    if input_rec.Headers["x-is-web-app"] == "1" or input_rec.Headers["x-is-web-app"] == "true" then
        set_out_prop(out, "IsAPI", false)
    else
        set_out_prop(out, "IsAPI", true)
    end
    if is_after_datetime(out, "2025-06-05T13:00:00") then
        datetime_add_minutes(out, -120)
    end
    return out
end

