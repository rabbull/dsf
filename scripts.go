package dsf

const (
	_SetNXOrGet = `
		if redis.pcall('SET', KEYS[1], ARGV[1], 'EX', ARGV[2], 'NX')
		then
			return {1, ""}
		else
			return {0, redis.pcall('GET', KEYS[1])}
		end
	`
	_DelIfEq = `
		local cur = redis.pcall('GET', KEYS[1])
		if cur == ARGV[1]
		then
			return redis.pcall('DEL', KEYS[1])
		else
			return -1
		end
	`
)
