package golang.java;

import java.sql.*;
public class GOComm {
	static public void safeClose(java.sql.Connection conn){
		try{
			if(conn != null && !conn.isClosed()){
				conn.close();
			}
		}catch(Exception e){}
	}
	static public void safeClose(java.sql.Statement stmt){
		try{
			if(stmt != null && !stmt.isClosed()){
				stmt.close();
			}
		}catch(Exception e){}
	}
	static public void safeClose(java.sql.PreparedStatement stmt){
		try{
			if(stmt != null && !stmt.isClosed()){
				stmt.close();
			}
		}catch(Exception e){}
	}
	static public void safeClose(java.sql.ResultSet rst){
		try{
			if(rst != null && !rst.isClosed()){
				rst.close();
			}
		}catch(Exception e){}
	}
	static String getError(Exception e){
		String mReturn = "GOJDBC:";
		if(e == null){
			mReturn += "No Error Or Unknown Error";
		}else{
			if(e instanceof SQLException){
				SQLException mSqlErr = (SQLException)e;
				String 		 mState = mSqlErr.getSQLState();
				mReturn += "["+mState+"]"+e.getMessage();
			}else{
				mReturn += e.getMessage();
			}
		}
		return mReturn;
	}
}
