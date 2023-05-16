local core = require "wind.core"
local serialize = require "wind.serialize"

local THREAD_MAIN <const> = 0
local THREAD_LOGGER <const> = 1
local THREAD_ROOT <const> = 2

local wind = {
    stateclass = {},        -- preload
    serviceclass = {},      -- preload
}


function wind.self()
    if not wind._self then
        local id, efd = core.self()
        wind._self = {
            id = id,
            efd = efd,
        }
    end
    return wind._self
end

function wind.send(thread_id, ...)
    -- wind.log("send", thread_id, ...)
    return core.send(thread_id, serialize.pack(wind.self().id, ...))
end

function wind.log(...)
    core.send(THREAD_LOGGER, serialize.pack(wind.self().id, "_log", ...))
end


function wind.error(...)
    core.send(THREAD_LOGGER, serialize.pack(wind.self().id, "_error", ...))
end


function wind.recv()
    local data = core.recv()
    if data then
        return serialize.unpack(data)
    end
end

wind.time = core.time


return wind