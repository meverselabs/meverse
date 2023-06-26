# Tracer 사용 예제

## traceCall

1. Struct/opcode logger
   ```
   {
       "jsonrpc":"2.0",
       "method":"debug_traceCall",
       "params":[
           {
               "from" : "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
               "to" : "0x5FbDB2315678afecb367f032d93F642f64180aa3",
               "data" : "0xa9059cbb00000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c800000000000000000000000000000000000000000000000000000000000004d2"
           },
           "latest",
           {"enableMemory": true, "disableStack":false, "disableStorage":false, "enableReturnData": true}
       ],
       "id":83
   }
   ```
2. callTracer, flatCallTrancer
   ```
    {
    "jsonrpc":"2.0",
    "method":"debug_traceCall",
    "params":[
        {
            "from" : "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
            "to" : "0x5FbDB2315678afecb367f032d93F642f64180aa3",
            "data" : "0xa9059cbb00000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c800000000000000000000000000000000000000000000000000000000000004d2"
        },
        "latest",
        { "tracer" : "callTracer", "tracerConfig" : { "withLog" : true} }
    ],
    "id":83
   }
   ```
3. 4byteTracer
   ```
   {
    "jsonrpc":"2.0",
    "method":"debug_traceCall",
    "params":[
        {
            "from" : "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
            "to" : "0x5FbDB2315678afecb367f032d93F642f64180aa3",
            "data" : "0xa9059cbb00000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c800000000000000000000000000000000000000000000000000000000000004d2"
        },
        "latest",
        { "tracer" : "4byteTracer"}
    ],
    "id":83
   }
   ```
4. prestate Tracer

   ```
   {
    "jsonrpc":"2.0",
    "method":"debug_traceCall",
    "params":[
        {
            "from" : "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
            "to" : "0x5FbDB2315678afecb367f032d93F642f64180aa3",
            "data" : "0xa9059cbb00000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c800000000000000000000000000000000000000000000000000000000000004d2"
        },
        "latest",
        { "tracer" : "prestateTracer", "tracerConfig" : { "diffMode" : true} }
    ],
    "id":83
    }
   ```

5. muxTracer
   ```
   {
    "jsonrpc":"2.0",
    "method":"debug_traceCall",
    "params":[
        {
            "from" : "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
            "to" : "0x5FbDB2315678afecb367f032d93F642f64180aa3",
            "data" : "0xa9059cbb00000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c800000000000000000000000000000000000000000000000000000000000004d2"
        },
        "latest",
        { "tracer" : "muxTracer", "tracerConfig":
            {
                "callTracer" :  { "withLog" : true},
                "4byteTracer" : {},
                "prestateTracer" : {"diffMode" : true}
            }
        }
    ],
    "id":83
   }
   ```

#tractTransaction

1. Struct/opcode logger
   ```
   {
       "jsonrpc":"2.0",
       "method":"debug_traceTransaction",
       "params":[
           "0xbf0ff9eee7b413361d42f6bc850ea39bc40f001b238a9ac2e2c6a4d1d3fe3756",
           {"enableMemory": true, "disableStack":false, "disableStorage":false, "enableReturnData": true}
       ],
       "id":83
   }
   ```
2. callTracer, flatCallTrancer
   ```
    {
    "jsonrpc":"2.0",
    "method":"debug_traceTransaction",
    "params":[
        "0xbf0ff9eee7b413361d42f6bc850ea39bc40f001b238a9ac2e2c6a4d1d3fe3756",
        { "tracer" : "callTracer", "tracerConfig" : { "withLog" : true} }
    ],
    "id":83
   }
   ```
3. 4byteTracer
   ```
   {
    "jsonrpc":"2.0",
    "method":"debug_traceTransaction",
    "params":[
        "0xbf0ff9eee7b413361d42f6bc850ea39bc40f001b238a9ac2e2c6a4d1d3fe3756",
        { "tracer" : "4byteTracer"}
    ],
    "id":83
   }
   ```
4. prestate Tracer

   ```
   {
    "jsonrpc":"2.0",
    "method":"debug_traceTransaction",
    "params":[
        "0xbf0ff9eee7b413361d42f6bc850ea39bc40f001b238a9ac2e2c6a4d1d3fe3756",
        { "tracer" : "prestateTracer", "tracerConfig" : { "diffMode" : true} }
    ],
    "id":83
    }
   ```

5. muxTracer
   ```
   {
    "jsonrpc":"2.0",
    "method":"debug_traceTransaction",
    "params":[
        "0xbf0ff9eee7b413361d42f6bc850ea39bc40f001b238a9ac2e2c6a4d1d3fe3756",
        { "tracer" : "muxTracer", "tracerConfig":
            {
                "callTracer" :  { "withLog" : true},
                "4byteTracer" : {},
                "prestateTracer" : {"diffMode" : true}
            }
        }
    ],
    "id":83
   }
   ```
