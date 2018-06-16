package golang.java;

import java.sql.*;
public class GOStatement {
	private PreparedStatement 	stmt;
	private Exception			except; //错误信息
	private int 				nParamCount;
	private int 				nTimeoutSecond; //超时时间
	public 	GOStatement(PreparedStatement s,int nTimeout){
		nParamCount = -1;
		stmt 		= s;
		nTimeoutSecond = nTimeout;
		ParameterMetaData pmd = null;
		try{
			pmd = stmt.getParameterMetaData();
			nParamCount = pmd.getParameterCount();	
		}catch(Exception e){
			nParamCount = -1;
		}
	}
	public void Close(){
		GOComm.safeClose(stmt);
		stmt = null;
	}
	private void SetError(Exception e){
		except = e;
	}
	public String GetError(){
		return GOComm.getError(except);
	}
	public int NumInput(){
		return nParamCount;
	}
	public int Execute(Object []args){
		try{
			if(args != null){
				for(int i=0;i<args.length;i++){
					stmt.setObject(i+1,args[i]);
				}
			}
			if(nTimeoutSecond > 0){
				stmt.setQueryTimeout(nTimeoutSecond);
			}
			int n = stmt.executeUpdate();
			return n;
		}catch(Exception e){
			SetError(e);
			return -1;
		}
	}
	public Object Query(Object []args){
		ResultSet 	rst  	= null;
		GORecordSet gorst 	= null;
		try {
			if(args != null){
				for(int i=0;i<args.length;i++){
					stmt.setObject(i+1,args[i]);
				}
			}
			if(nTimeoutSecond > 0){
				stmt.setQueryTimeout(nTimeoutSecond);
			}
			rst = stmt.executeQuery();
			gorst = new GORecordSet(rst);
			gorst.initResult();
		} catch (Exception e) {
			SetError(e);
			GOComm.safeClose(rst);
			return null;
		}
		return gorst;
	}
}
