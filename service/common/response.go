package common

type CommonRsp struct {
	ErrorCode int `json:"errorCode"`
	Message string `json:"message"`
	Error bool `json:"error"`
	Result interface{} `json:"result"`
	Params map[string]interface{} `json:"params"`
}

type CommonError struct {
	ErrorCode int `json:"errorCode"`
	Params map[string]interface{} `json:"params"`
}

const (
	ResultSuccess = 10000000
	ResultWrongRequest = 10000001
	ResultCreateDirError=10000027
	ResultBase64DecodeError=10000028
	ResultCreateFileError=10000029
	ResultCannotLoginCRV = 10100002
	ResultNoParams = 10100003
	ResultNoDtcList = 10100004
	ResultNoDiagConf = 10100005
	ResultNoNoVehicle = 10100006
	ResultQueryRequestError = 10100007
	ResultMqttClientError = 10100008
	ResultSaveDataError = 10100010
	ResultCacheSendRecError = 10100011
	ResultRepeatedEcu = 10100012
	ResultMultiProject = 10100013
	ResultParamWithoutEcu = 10100014
	ResultWrongDiagConf = 10100015
	ResultJonsMarshalError = 10000043
)

var errMsg = map[int]CommonRsp{
	ResultSuccess:CommonRsp{
		ErrorCode:ResultSuccess,
		Message:"操作成功",
		Error:false,
	},
	ResultJonsMarshalError:CommonRsp{
		ErrorCode:ResultJonsMarshalError,
		Message:"将对象转换为JSON文本时发生错误，请与管理员联系处理",
		Error:true,
	},
	ResultWrongRequest:CommonRsp{
		ErrorCode:ResultWrongRequest,
		Message:"请求参数错误，请检查参数是否完整，参数格式是否正确",
		Error:true,
	},
	ResultCannotLoginCRV:CommonRsp{
		ErrorCode:ResultWrongRequest,
		Message:"连接到基础数据平台失败，请与管理员联系处理",
		Error:true,
	},
	ResultNoParams:CommonRsp{
		ErrorCode:ResultNoParams,
		Message:"下发参数时未查询到对应的配置参数信息，请刷新页面数据后重新尝试",
		Error:true,
	},
	ResultNoDtcList:CommonRsp{
		ErrorCode:ResultNoDtcList,
		Message:"下发参数时未查询到对应的DTC信息，请确认DTC信息配置正确后重新尝试",
		Error:true,
	},
	ResultNoDiagConf:CommonRsp{
		ErrorCode:ResultNoDiagConf,
		Message:"下发参数时未查询到诊断配置信息，请确认诊断信息正确配置后重新尝试",
		Error:true,
	},
	ResultNoNoVehicle:CommonRsp{
		ErrorCode:ResultNoNoVehicle,
		Message:"下发参数时未查询到车辆信息，请确认诊车辆信息正确配置后重新尝试",
		Error:true,
	},
	ResultQueryRequestError:CommonRsp{
		ErrorCode:ResultQueryRequestError,
		Message:"下发参数时发送查询参数请求失败，请与管理员联系处理",
		Error:true,
	},
	ResultMqttClientError:CommonRsp{
		ErrorCode:ResultMqttClientError,
		Message:"下发参数时连接MQTT失败，请与管理员联系处理",
		Error:true,
	},
	ResultSaveDataError:CommonRsp{
		ErrorCode:ResultSaveDataError,
		Message:"保存数据到数据时发生错误，请与管理员联系处理",
		Error:true,
	},
	ResultCacheSendRecError:CommonRsp{
		ErrorCode:ResultCacheSendRecError,
		Message:"缓存下发参数时发生错误，请与管理员联系处理",
		Error:true,
	},
	ResultRepeatedEcu:CommonRsp{
		ErrorCode:ResultRepeatedEcu,
		Message:"下发参数中包含重复的ECU，请检查选择参数正确后重新尝试",
		Error:true,
	},
	ResultMultiProject:CommonRsp{
		ErrorCode:ResultMultiProject,
		Message:"下发参数中不能包含多个项目的参数，请检查选择参数正确后重新尝试",
		Error:true,
	},
	ResultParamWithoutEcu:CommonRsp{
		ErrorCode:ResultParamWithoutEcu,
		Message:"下发参数中未指定ECU信息，请检查选择参数正确后重新尝试",
		Error:true,
	},
	ResultCreateDirError:CommonRsp{
		ErrorCode:ResultCreateDirError,
		Message:"保存文件时创建文件夹失败，请与管理员联系处理",
		Error:true,
	},
	ResultBase64DecodeError:CommonRsp{
		ErrorCode:ResultBase64DecodeError,
		Message:"保存文件时文件内容Base64解码失败，请与管理员联系处理",
		Error:true,
	},
	ResultCreateFileError:CommonRsp{
		ErrorCode:ResultBase64DecodeError,
		Message:"创建文件失败，请与管理员联系处理",
		Error:true,
	},
	ResultWrongDiagConf:CommonRsp{
		ErrorCode:ResultWrongDiagConf,
		Message:"下发参数时诊断配置信息不完整，请确认诊断信息正确配置后重新尝试",
		Error:true,
	},
}

func CreateResponse(err *CommonError,result interface{})(*CommonRsp){
	if err==nil {
		commonRsp:=errMsg[ResultSuccess]
		commonRsp.Result=result
		return &commonRsp
	}

	commonRsp:=errMsg[err.ErrorCode]
	commonRsp.Result=result
	commonRsp.Params=err.Params
	return &commonRsp
}

func CreateError(errorCode int,params map[string]interface{})(*CommonError){
	return &CommonError{
		ErrorCode:errorCode,
		Params:params,
	}
}