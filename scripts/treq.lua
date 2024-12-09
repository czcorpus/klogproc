function preprocess (input_rec, buffer)
    return {input_rec}
end

function transform (input_rec)
    local out = transform_default(input_rec, 0)
    if input_rec.IsWebApp then
        set_out_prop(out, "IsAPI", false)
    else
        set_out_prop(out, "IsAPI", true)
    end
    return out
end