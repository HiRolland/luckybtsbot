local http = require("http");
local json = require("json");

local account = nil
local btssymbol = 'TEST'
local url = 'http://127.0.0.1:18080/'

-- JSON RPC 客户端
local rpc = {}
rpc.call = function(method, params)
    local req = {jsonrpc= "2.0", method= method, params= params, id= 1}
    local jstring, err = json:dump(req)
    if err ~= nil then
        return nil, err
    end

    print(string.format('[Lua script] json rpc2 request: %s', jstring))
    local resp, err = http:post(url, 'application/json', jstring)
    if err ~= nil then
        print(string.format('[Lua script] json rpc2 request error: %s', err))
        return nil, err
    end
    print(string.format('[Lua script] json rpc2 respone status_code: %s',
        resp['status_code']))

    print(string.format('[Lua script] json rpc2 respone: %s', resp['body']))
    local result, err = json:parse(resp['body'])
    if err ~= nil then
        return nil, err
    end

    if result['error'] ~= nil then
        local err = result['error']
        print(string.format('[Lua script] json rpc2 respone error, code: %d, %s',
            err['code'], err['message']))
        return nil, string.format('code: %d, %s', err['code'], err['message'])
    end
    return result['result'], nil
end

-- 获取账户
function get_account()
    if account == nil then
        local to, err = rpc.call('account', {})
        if err ~= nil then
            return ''
        end
        account = to
    end
    return account
end

-- 时钟事件
-- @param delaytime <number>
function on_tick(delaytime)
end

-- 账户是否有效
-- @param address <string> 地址
-- @return <boolean>
function valid_address(address)
    if string.len(address) < 3 then
        return false
    end
    if string.len(address) >= 64 then
        return false
    end
    for c in string.gmatch(address, ".") do
        local valid = false
        if c == '-' then
            valid = true
        end
        if c >= 'a' and c <= 'z' then
            valid = true
        end
        if c >= '0' and c <= '9' then
            valid = true
        end
        if valid ~= true then
            return false
        end
    end
    return true
end

-- 获取充值地址
-- @param userid <string> 用户ID
-- @return address <string>
-- @return memo <string or nil>
function deposit_address(userid)
    return get_account(), userid
end

-- 接收提现请求
-- @param to <string> 目标地址
-- @param symbol <string> 货币符号
-- @param amount <string> 提现金额
-- @param future <Future> 处理完成必须调用set_result(txid, error)方法
function on_withdraw(to, symbol, amount, future)
    if symbol ~= btssymbol then
        print('[Lua script] withdraw fail, invalid symbol')
        future:set_result(nil, 'invalid symbol')
        return
    end
    
    local txid, err = rpc.call('transfer', {to, symbol, amount, ''})
    if err ~= nil then
        print(string.format('[Lua script] withdraw fail, %s', err))
        future:set_result(nil, err)
        return
    end  
    print(string.format('[Lua script] withdraw success, txid: %s', txid))
    future:set_result(txid, nil)
end

-- 交易是否有效
-- @param txid <string> 交易ID
-- @param from <string> 来源地址
-- @param to <string> 目标地址
-- @param symbol <string> 货币符号
-- @param amount <string> 交易金额
-- @param memo <string> 备注信息
-- @return <boolean>
function valid_transaction(txid, from, to, symbol, amount, memo)
    if to ~= get_account() then
        return false
    end
    return true
end
