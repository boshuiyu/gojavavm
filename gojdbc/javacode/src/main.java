import golang.java.*;
public class main {
	/*
	 * 主程序
	 * */
	public static void main(String[] args) {
		System.out.println("开始");
		GOConnect conn = new GOConnect();
		boolean b = conn.Open("jdbc:mysql://172.16.6.98:3306/mydbtest", "root", "yw@123456",4);
		GOStatement stmt = (GOStatement)conn.Prepare("select sleep(6)");
		stmt.Query(null);
		stmt.Close();
		conn.Close();
		
	}
}
