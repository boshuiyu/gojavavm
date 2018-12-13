package golang.java;

import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.SQLException;

public class GOConnect {
	static 
    {
        try { Class.forName("com.mysql.jdbc.Driver");}catch(Exception e ){
        	e.printStackTrace();
        }
        try { Class.forName("com.microsoft.sqlserver.jdbc.SQLServerDriver");}catch(Exception e ){
        	e.printStackTrace();
        }
        try { Class.forName("dm.jdbc.driver.DmDriver");}catch(Exception e ){
        	e.printStackTrace();
        }
        try { Class.forName("com.kingbase.Driver");}catch(Exception e ){
        	e.printStackTrace();
        }
    }
	//函数定义
	private java.sql.Connection  	conn; 		//数据库连接
	private Exception				except; 	//异常信息
	private int					nTimeout; 	//全局的超时时间(秒)
	//构造函数
	public GOConnect(){
		conn = null;
		except = null;
		nTimeout = 0;
	}
	//重置错误信息
	private void SetError(Exception e){
		except = e;
	}
	public String GetError(){
		return GOComm.getError(except);
	}
	//打开数据库连接
	public boolean Open(String url,String user,String pass,long nQueryTimeout){
		Close();
		if(nQueryTimeout >= 0){
			nTimeout = (int)nQueryTimeout;
		}
		try{
			conn = DriverManager.getConnection(url, user, pass);
			conn.setAutoCommit(true);
		}catch(Exception e){
			SetError(e);
			Close();
			return false;
		}
		return true;
	}
	public void Close(){
		if(conn != null){
			GOComm.safeClose(conn);
			conn = null;
		}
	}
	public Object Prepare(String query){
		GOStatement 		res = null;
		PreparedStatement 	stmt = null;
		try{
			stmt =conn.prepareStatement(query);
			res = new GOStatement(stmt,nTimeout);
		}catch(Exception e){
			SetError(e);
			res = null;
		}
		return res;
	}
	
	//事务相关
	public boolean Begin(){
		try {
			conn.setAutoCommit(false);
		} catch (SQLException e) {
			SetError(e);
			try{conn.setAutoCommit(true);}catch(Exception e1){}
			return false;
		}
		return true;
	}
	public boolean Commit(){
		try{
			conn.commit();
		}catch (SQLException e) {
			SetError(e);
			return false;
		}finally{
			try{conn.setAutoCommit(true);}catch(Exception e){}
		}
		return true;
	}
	public boolean Rollback(){
		try{
			conn.rollback();
		}catch (SQLException e) {
			SetError(e);
			return false;
		}finally{
			try{conn.setAutoCommit(true);}catch(Exception e){}
		}
		return true;
	}
}
