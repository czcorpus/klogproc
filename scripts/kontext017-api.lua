function transform(input_rec)
    local out = transform_default(input_rec, 0)
    if input_rec.Request.HTTPIsWebApp == "1" and input_rec.IsAPI then
	    set_out_prop(out, "IsAPI", false)
    end
    return out
end

