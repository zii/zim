package redis

import "github.com/gomodule/redigo/redis"

// Only increase the value if the key exists.
var lua_incrby = redis.NewScript(1, `
	local key = KEYS[1];
	local amount = tonumber(ARGV[1]);
	if amount < 0 then
		local n = redis.call("GET", key);
		n = tonumber(n or 0);
		if n < -amount then
			return nil;
		else
			n = redis.call("INCRBY", key, amount);
			return n;
		end
	elseif redis.call("EXISTS", key) == 1 then
		local n = redis.call("INCRBY", key, amount);
		return n;
	end
	return nil;
`)

// set if higher
// if not exists return OK, if updated return the increment, if not updated return 0
var lua_sethi = redis.NewScript(1, `
	local c = tonumber(redis.call("GET", KEYS[1]));
	if c then
		local a = tonumber(ARGV[1]);
		if a > c then
			redis.call("SET", KEYS[1], ARGV[1]);
			return a - c;
		else
			return 0;
		end
	else
		return redis.call("SET", KEYS[1], ARGV[1]);
	end
`)
