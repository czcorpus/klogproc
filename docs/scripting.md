# Scripting Klogproc with Lua

For some application logs, Klogproc allows customization of their log
processing without recompiling the whole application. The key principle
is that a user defines two functions which are then applied repeatedly
to each processed log record of the specified type.

## Input record


The input record provides access to its attributes in the same way as in the Go language - i.e. the attributes use camel case and start with an uppercase letter.

As each application may provide different date and time encodings, input record has a `GetTime` method that returns the RFC3339 encoded date and time.


## Output record

Output record represents a normalized record shared by all logged applications. Klogproc typically provides a default way  of converting the application's own input log into the output format. If a Lua
script is configured for the application, Klogproc will only call the Lua-defined transformation function which means that the default conversion is omitted. For use cases where the default conversion is still required and the purpose of the Lua script is just to customise the default conversion, it can be called explicitly:

```lua
local out = transform_default(input_rec)
-- modify the output
-- ...
```

For cases where a new empty record is needed for the script, just use:

```lua
local out = new_out_record()
-- set output properties
-- ...
```

To set a property in output record, Klogproc requires using `set_out_prop` function:

```lua
set_out_prop(out, name, value)
```

In case the attribute cannot be set (typically because it does not exist),
the script ends with and error.

To test whether a record (input or output) has a property:

```lua
if record_prop_exists(any_rec, name, value) then
    -- we're sure the attribute `name` can be set
end
```

Once the output record is set, it is necessary to set an ID that will be used as the database ID. Klogproc prefers deterministic IDs, which allow repeated data import without duplicating data with the same content but different IDs. To obtain a deterministic ID for an output record, use:

```lua
local id = out_rec_deterministic_id(out)
```

The function can be called repeatedly and for the same attributes (the ID itself is always ignored), it will return the same ID (hash).


## Debugging, logging

For printing contents of a value, use:

```lua
dump = require('dump')
print(dump(my_value))
```

For logging:

```lua
log.info("message", map_with_args)
-- other available logging levels are: "warn", "debug", "error"
```
The second argument is optional.


## Global variables

* `app_type` - an id representing a logged application (`kontext`, `wag`, `korpusdb` etc.),
* `app_version` - a string representing a variant of the application; for some applications, versions are not defined,
* `anonymous_users` - a list of CNC database IDs defined in Klogproc configuration

## Function preprocess()

The preprocess function is called before an input record is transformed into
an output record. Its purpose is to provide the following options:

1. decide whether to process the input_rec at all
   (just return `{}` to skip the record)
1. For applications, where a "query request" is hard to define (e.g. `mapka`),
   it allows to generate "virtual" input records that are somehow derived from the real ones. E.g. in the `mapka v3` application, we search for
   activity clusters and for each input cluster we generate a single record.

Since all the application transformers written in Go require the `preprocess`
function to be defined, in Lua it is possible to call it manually as
Klogproc will not do it once a Lua script is defined. To do this, use
function `preprocess_default(input_rec, buffer)`. In case no preprocessing
is required, simply return the original record in a table:

```lua
function preprocess(input_rec, buffer)
    return {input_rec}
end
```

### Buffer access

`TODO`

## Function transform()

The `transform` function converts an input record into a normalized form.
As mentioned above, if a Lua script is defined for an application, Klogproc will not automatically call the hardcoded version of the transform function. So if it is needed, it has to be called explicitly.

```lua
-- transform function processes the input record and returns an output record
function transform(input_rec)
    local out = transform_default(input_rec)
    -- now we modify the Path property already set by transform_default
	set_out_prop(
        out, "Path", string.format("%s/modified/value", input_rec.Path))
    return out
end
```


## Function datetime_add_minutes(num_min)

The `datetime_add_minutes` allows for shifting log record time forwards and backwards
to correct possible timezone issues.

```lua
function transform(input_rec)
    local out = transform_default(input_rec)
    datetime_add_minutes(out, -120)
    return out
end
```

## Function is_before_datetime(rec, datetime)

The `is_before_datetime` tests whether the `rec` record has its datetime property before
the provided `datetime` argument. The format is `2006-01-02T15:04:05-07:00`.

## Function is_after_datetime(rec, datetime)

The `is_after_datetime` is analogous to the `is_before_datetime`.