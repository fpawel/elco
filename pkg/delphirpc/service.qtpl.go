// This file is automatically generated by qtc from "service.qtpl".
// See https://github.com/valyala/quicktemplate for details.

//line service.qtpl:1
package delphirpc

//line service.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line service.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line service.qtpl:3
func (x *ServicesSrc) StreamUnit(qw422016 *qt422016.Writer) {
	//line service.qtpl:3
	qw422016.N().S(` 
unit `)
	//line service.qtpl:4
	qw422016.E().S(x.unitName)
	//line service.qtpl:4
	qw422016.N().S(`;

interface

`)
	//line service.qtpl:8
	qw422016.E().S(uses(x.interfaceUses))
	//line service.qtpl:8
	qw422016.N().S(`

type 
    `)
	//line service.qtpl:11
	for _, srv := range x.services {
		//line service.qtpl:11
		qw422016.N().S(` T`)
		//line service.qtpl:11
		qw422016.E().S(srv.serviceName)
		//line service.qtpl:11
		qw422016.N().S(` = class 
    public
        `)
		//line service.qtpl:13
		for _, m := range srv.methods {
			//line service.qtpl:13
			qw422016.N().S(m.signature(""))
			//line service.qtpl:13
			qw422016.N().S(`;static;
        `)
			//line service.qtpl:14
		}
		//line service.qtpl:14
		qw422016.N().S(` 
    end;`)
		//line service.qtpl:15
	}
	//line service.qtpl:15
	qw422016.N().S(`

implementation 

`)
	//line service.qtpl:19
	qw422016.E().S(uses(x.implUses))
	//line service.qtpl:19
	qw422016.N().S(`

`)
	//line service.qtpl:21
	for _, srv := range x.services {
		//line service.qtpl:21
		qw422016.N().S(` `)
		//line service.qtpl:21
		for _, m := range srv.methods {
			//line service.qtpl:21
			qw422016.N().S(m.body(srv.serviceName))
			//line service.qtpl:21
			qw422016.N().S(`
`)
			//line service.qtpl:22
		}
		//line service.qtpl:22
	}
	//line service.qtpl:22
	qw422016.N().S(` 
end.`)
//line service.qtpl:23
}

//line service.qtpl:23
func (x *ServicesSrc) WriteUnit(qq422016 qtio422016.Writer) {
	//line service.qtpl:23
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line service.qtpl:23
	x.StreamUnit(qw422016)
	//line service.qtpl:23
	qt422016.ReleaseWriter(qw422016)
//line service.qtpl:23
}

//line service.qtpl:23
func (x *ServicesSrc) Unit() string {
	//line service.qtpl:23
	qb422016 := qt422016.AcquireByteBuffer()
	//line service.qtpl:23
	x.WriteUnit(qb422016)
	//line service.qtpl:23
	qs422016 := string(qb422016.B)
	//line service.qtpl:23
	qt422016.ReleaseByteBuffer(qb422016)
	//line service.qtpl:23
	return qs422016
//line service.qtpl:23
}

//line service.qtpl:25
func (x method) streambody(qw422016 *qt422016.Writer, srvName string) {
	//line service.qtpl:25
	qw422016.N().S(` 
`)
	//line service.qtpl:26
	qw422016.N().S(x.signature("T" + srvName + "."))
	//line service.qtpl:26
	qw422016.N().S(`;
var
    req : ISuperobject;
begin
    req := `)
	//line service.qtpl:30
	if x.namedParams {
		//line service.qtpl:30
		qw422016.N().S(`SO`)
		//line service.qtpl:30
	} else {
		//line service.qtpl:30
		qw422016.N().S(`SA([])`)
		//line service.qtpl:30
	}
	//line service.qtpl:30
	qw422016.N().S(`;
    `)
	//line service.qtpl:31
	for _, p := range x.params {
		//line service.qtpl:31
		qw422016.N().S(x.genSetParam(p))
		//line service.qtpl:31
		qw422016.N().S(`;
    `)
		//line service.qtpl:32
	}
	//line service.qtpl:33
	if x.procedure {
		//line service.qtpl:33
		qw422016.N().S(`ThttpRpcClient.GetResponse(`)
		//line service.qtpl:34
		qw422016.N().S(x.remoteMethod(srvName))
		//line service.qtpl:34
		qw422016.N().S(`, req); `)
		//line service.qtpl:35
	} else {
		//line service.qtpl:36
		if x.retPODType && !x.retArray {
			//line service.qtpl:36
			qw422016.N().S(`SuperObject_Get(ThttpRpcClient.GetResponse(`)
			//line service.qtpl:37
			qw422016.N().S(x.remoteMethod(srvName))
			//line service.qtpl:37
			qw422016.N().S(`, req), Result); `)
			//line service.qtpl:38
		} else {
			//line service.qtpl:38
			qw422016.N().S(`ThttpRpcClient.Call(`)
			//line service.qtpl:39
			qw422016.N().S(x.remoteMethod(srvName))
			//line service.qtpl:39
			qw422016.N().S(`, req, Result); `)
			//line service.qtpl:40
		}
		//line service.qtpl:41
	}
	//line service.qtpl:42
	qw422016.N().S(`
end;
`)
//line service.qtpl:44
}

//line service.qtpl:44
func (x method) writebody(qq422016 qtio422016.Writer, srvName string) {
	//line service.qtpl:44
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line service.qtpl:44
	x.streambody(qw422016, srvName)
	//line service.qtpl:44
	qt422016.ReleaseWriter(qw422016)
//line service.qtpl:44
}

//line service.qtpl:44
func (x method) body(srvName string) string {
	//line service.qtpl:44
	qb422016 := qt422016.AcquireByteBuffer()
	//line service.qtpl:44
	x.writebody(qb422016, srvName)
	//line service.qtpl:44
	qs422016 := string(qb422016.B)
	//line service.qtpl:44
	qt422016.ReleaseByteBuffer(qb422016)
	//line service.qtpl:44
	return qs422016
//line service.qtpl:44
}

//line service.qtpl:46
func (x method) streamgenSetParam(qw422016 *qt422016.Writer, p param) {
	//line service.qtpl:47
	if x.namedParams {
		//line service.qtpl:48
		qw422016.N().S(p.setFieldInstruction())
		//line service.qtpl:49
	} else {
		//line service.qtpl:49
		qw422016.N().S(`req.AsArray.Add(`)
		//line service.qtpl:50
		qw422016.E().S(p.name)
		//line service.qtpl:50
		qw422016.N().S(`) `)
		//line service.qtpl:51
	}
//line service.qtpl:52
}

//line service.qtpl:52
func (x method) writegenSetParam(qq422016 qtio422016.Writer, p param) {
	//line service.qtpl:52
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line service.qtpl:52
	x.streamgenSetParam(qw422016, p)
	//line service.qtpl:52
	qt422016.ReleaseWriter(qw422016)
//line service.qtpl:52
}

//line service.qtpl:52
func (x method) genSetParam(p param) string {
	//line service.qtpl:52
	qb422016 := qt422016.AcquireByteBuffer()
	//line service.qtpl:52
	x.writegenSetParam(qb422016, p)
	//line service.qtpl:52
	qs422016 := string(qb422016.B)
	//line service.qtpl:52
	qt422016.ReleaseByteBuffer(qb422016)
	//line service.qtpl:52
	return qs422016
//line service.qtpl:52
}

//line service.qtpl:55
func (x method) streamsignature(qw422016 *qt422016.Writer, headPart string) {
	//line service.qtpl:55
	qw422016.N().S(`class `)
	//line service.qtpl:57
	if x.procedure {
		//line service.qtpl:57
		qw422016.N().S(`procedure `)
		//line service.qtpl:59
	} else {
		//line service.qtpl:59
		qw422016.N().S(`function `)
		//line service.qtpl:61
	}
	//line service.qtpl:62
	qw422016.E().S(headPart)
	//line service.qtpl:62
	qw422016.E().S(x.methodName)
	//line service.qtpl:63
	if len(x.params) > 0 {
		//line service.qtpl:63
		qw422016.N().S(`( `)
		//line service.qtpl:65
		for i, p := range x.params {
			//line service.qtpl:66
			qw422016.E().S(p.name)
			//line service.qtpl:66
			qw422016.N().S(`: `)
			//line service.qtpl:66
			qw422016.N().S(p.String())
			//line service.qtpl:67
			if i < len(x.params)-1 {
				//line service.qtpl:67
				qw422016.N().S(`; `)
				//line service.qtpl:67
			}
			//line service.qtpl:68
		}
		//line service.qtpl:68
		qw422016.N().S(`) `)
		//line service.qtpl:70
	}
	//line service.qtpl:72
	if !x.procedure {
		//line service.qtpl:72
		qw422016.N().S(`: `)
		//line service.qtpl:74
		if x.retArray {
			//line service.qtpl:75
			qw422016.N().S("TArray<")
			//line service.qtpl:75
			qw422016.N().S(x.retDelphiType)
			//line service.qtpl:75
			qw422016.N().S(">")
			//line service.qtpl:76
		} else {
			//line service.qtpl:77
			qw422016.N().S(x.retDelphiType)
			//line service.qtpl:78
		}
		//line service.qtpl:79
	}
//line service.qtpl:80
}

//line service.qtpl:80
func (x method) writesignature(qq422016 qtio422016.Writer, headPart string) {
	//line service.qtpl:80
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line service.qtpl:80
	x.streamsignature(qw422016, headPart)
	//line service.qtpl:80
	qt422016.ReleaseWriter(qw422016)
//line service.qtpl:80
}

//line service.qtpl:80
func (x method) signature(headPart string) string {
	//line service.qtpl:80
	qb422016 := qt422016.AcquireByteBuffer()
	//line service.qtpl:80
	x.writesignature(qb422016, headPart)
	//line service.qtpl:80
	qs422016 := string(qb422016.B)
	//line service.qtpl:80
	qt422016.ReleaseByteBuffer(qb422016)
	//line service.qtpl:80
	return qs422016
//line service.qtpl:80
}
