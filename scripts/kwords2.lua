function transform(input_rec)
    local out = transform_default(input_rec, 0)
    if input_rec.Headers["x-is-web-app"] == "1" or input_rec.Headers["x-is-web-app"] == "true" then
        set_out_prop(out, "IsAPI", false)
    else
        set_out_prop(out, "IsAPI", true)
    end
    return out
end

