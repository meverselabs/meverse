var STRIP_COMMENTS = /(\/\/.*$)|(\/\*[\s\S]*?\*\/)|(\s*=[^,\)]*(('(?:\\'|[^'\r\n])*')|("(?:\\"|[^"\r\n])*"))|(\s*=[^,\)]*))/mg;
var ARGUMENT_NAMES = /([^\s,]+)/g;

function getParamNames(func) {
    var fnStr = func.toString().replace(STRIP_COMMENTS, '');
    var result = fnStr.slice(fnStr.indexOf('(')+1, fnStr.indexOf(')')).match(ARGUMENT_NAMES);
    if(result === null)
        result = [];
    return result;
}

function $args(func) {
    let funcStr = ""
    if (typeof func == "function") {
        funcStr = Function.toString.call(func)
    } else if (typeof func == "string") {
        funcStr = func
    } else {
        return null
    }
    return funcStr
        .replace(/[\s]*=[\s]*[\x22\x27\x60](.*?)*[\x22\x27\x60]/g, "")
        .replace(/[/][/].*$/mg,'')
        .replace(/\s+/g, '')
        .replace(/[/][*][^/*]*[*][/]/g, '')
        .split(/(\){)|(\)=>{)/, 1)[0].replace(/^[^(]*[(]/, '')
        .replace(/=[^,]+/g, '')
        .split(',').filter(Boolean);
}

function abi() {
	let list = [];
	for (var k in window) {
		if (window.hasOwnProperty(k) && k.indexOf("_") != 0 && !systemFuncList.includes(k)) {
			let f = window[k];
			if (typeof f == "function") {
				let args = $args(f)
                let params = []
                for (var i in args) {
                    params.push({"internalType": "string","name": args[i],"type": "string"})
                }
				list.push({
                    "inputs": params,
                    "name": f.name,
                    "outputs": [{"internalType": "string","name": "","type": "string"}],
                    "stateMutability": "nonpayable",
                    "type": "function"
                })
			}
		}
	}
	return list;
}

function $argsPrint(func) {
    let funcStr = ""
    if (typeof func == "function") {
        funcStr = Function.toString.call(func)
    } else if (typeof func == "string") {
        funcStr = func
    } else {
        console.log(func)
        return null
    }
    console.log(funcStr)
    funcStr = funcStr.replace(/[\s]*=[\s]*[\x22\x27\x60](.*?)*[\x22\x27\x60]/g, "")
    console.log("1", funcStr)
    funcStr = funcStr.replace(/[/][/].*$/mg,'')
    console.log("2", funcStr)
    funcStr = funcStr.replace(/\s+/g, '')
    console.log("3", funcStr)
    funcStr = funcStr.replace(/[/][*][^/*]*[*][/]/g, '')
    console.log("4", funcStr)
    funcStr = funcStr.split(/(\){)|(\)=>{)/, 1)[0].replace(/^[^(]*[(]/, '')
    console.log("5", funcStr)
    funcStr = funcStr.replace(/=[^,]+/g, '')
    console.log("6", funcStr)
    funcStr = funcStr.split(',').filter(Boolean);
    console.log("7", funcStr)
    
    return funcStr
}

function test(f, result) {
    let a = $args(f)
    if (a+"" != result+"" && a != result) {
        console.trace();
        $argsPrint(f)
        throw a+" is not match result("+result+")"
    }
}

test(getParamNames, ['func']) // returns ['func']
test(function (a,b,c,d){},  ['a','b','c','d']) // returns ['a','b','c','d']
test(function (a,/*b,c,*/d){}, ['a','d']) // returns ['a','d']
test(function (){}, []) // returns []
test(function (a=4*(5/3), b) {}, ['a','b']) // returns []

test(function (a,b,c){}, ['a','b','c'])
test(function named(a, b, c) {}, ['a','b','c'])
test(function (a /* = 1 */, b /* = true */) {}, ['a','b'])
test(function fprintf(handle, fmt /*, ...*/) {}['handle','fmt'])
test(function( a, b = 1, c ){}, ['a','b','c'])
test(function (a, // single-line comment xjunk) {}
b){}, ['a','b'])
test(function (a /* fooled you{})
*/
,v){}, ['a','v'])
test(function (a /* function() yes */, 
/* no, */b)/* omg! */{}, ['a','b'])
test(function ( A, b 
    ,c ,d 
     )
      {}, ['A','b', 'c', 'd'])
test(function (a,b){}, ['a','b'])
test(function $args(func) {}, ['func'])
test(null, undefined)
test(function Object() {}, [])
test(function Object(asdf = "//", etwqt) {}, ['asdf','etwqt'])
test(function Object(asdf = `//`, etwqt) {}, ['asdf','etwqt'])
test(function Object(asdf = '//', etwqt) {}, ['asdf','etwqt'])
test((asdf = '///**/', etwqt) => {}, ['asdf','etwqt'])
test((asdf = "///**/'", etwqt) => {}, ['asdf','etwqt'])
test((asdf = `///**/"`, etwqt) => {}, ['asdf','etwqt'])
test((asdf = `///**/"''`, etwqt) => {}, ['asdf','etwqt'])
test((asdf = `'"///**/"'`, etwqt) => {}, ['asdf','etwqt'])
