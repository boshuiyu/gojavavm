import golang.java.*;
public class main {
	/*
	 * 主程序
	 * */
	public static void main(String[] args) {
		System.out.println("调用参数如:java -jar gojdbc.jar jdbc:mysql://127.0.0.1:3306/testdb root 123456");
		if(args == null || args.length<3){
			System.out.println("启动参数错误");
			return;
		}
		System.out.println("测试连接数据库:"+args[0]);
		GOConnect conn = new GOConnect();
		boolean ok = conn.Open(args[0],args[1],args[2],4);
		if(!ok){
			System.out.println("连接数据库失败:"+conn.GetError());
			return;
		}
		String mSqlText = "SELECT 1 As TestValue"; 
		System.out.println("连接数据库成功,验证查询语句:"+mSqlText);
		GOStatement stmt = (GOStatement)conn.Prepare(mSqlText);
		Object rst=stmt.Query(null);
		System.out.println("查询结果:");
		System.out.println(rst);
		stmt.Close();
		conn.Close();
		System.out.println("执行完成");
	}
}
