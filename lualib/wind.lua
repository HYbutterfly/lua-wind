local core = require "wind.core"
local serialize = require "wind.serialize"
local config = require "config"

local wind = {}


do
    wind.id, wind.efd = core.self()
end


function wind.send(thread_id, ...)
    return core.send(thread_id, serialize.pack(...))
end

-- for main thread
function wind.newservice(worker, name, ...)
    wind.send(worker, "newservice", name, ...)
    for i = 1, config.nworker do
        if i ~= worker then
            wind.send(i, "sync_service_worker", name, worker)
        end
    end
end


function wind.recv()
    local data = core.recv()
    if data then
        return serialize.unpack(data)
    end
end

wind.time = core.time
wind.sleep = core.sleep


return wind