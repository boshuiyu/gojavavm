package golang.java;

import java.util.*;
import java.sql.*;
import java.text.*;
//用于生成JSON的结构
public class GORecordSet {
	private ResultSet 			rst; 		//结果集
	private Exception			except; 	//错误信息
	private int					nColumnSize;//字段数
	private List<String> 		lstColumn;	//字段列表
	private DateFormat 			dateFormat;
	public 	GORecordSet(ResultSet s){
		rst 		= s;
		except 		= null;
		lstColumn   = new ArrayList<String>();
		dateFormat  = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
	}
	protected void  initResult() throws SQLException{
		ResultSetMetaData md = rst.getMetaData();
		nColumnSize = md.getColumnCount();
		for(int i=0;i<nColumnSize;i++){
			String colName = md.getColumnName(i+1);
			if(colName == null || colName.length() == 0){
				colName = "$column_"+String.valueOf(i)+"$";
			}
			lstColumn.add(colName);
		}
		
	}
	private void SetError(Exception e){
		except = e;
	}
	public String GetError(){
		return GOComm.getError(except);
	}
	public void Close(){
		GOComm.safeClose(rst);
		rst = null;
	}
	public String Columns(){
		String mRes = "";
		for(int i=0;i<lstColumn.size();i++){
			if(mRes.length() == 0){
				mRes = lstColumn.get(i);
			}else{
				mRes += "<*>"+lstColumn.get(i);
			}
		}
		return mRes;
	}
	public int Next(){
		//返回1还有记录,-1无记录,0错误
		int nRet = 0;
		try {
			if(rst.next()){
				nRet = 1;
			}else{
				nRet = -1;
			}
		} catch (SQLException e) {
			SetError(e);
			nRet = 0;
		}
		return nRet;
	}
	
	public Object Fetch(){
		try {
			//没有数据了返回0个数据长度，有错误返回null
			if(!rst.next()){
				return new Object[0];
			}
			//获取当前条记录的各个数据并且转换为适合处理的对象
			Object[] arrObjs = new Object[nColumnSize];
			for(int i=0;i<nColumnSize;i++){
				Object obj = rst.getObject(i+1);
				arrObjs[i] = convertDataObject(obj);
			}
			return arrObjs;
		} catch (SQLException e) {
			//错误返回null
			SetError(e);
			return null;
		}
	}
	
	private Object convertDataObject(Object obj){
		if(obj == null){
			return obj;
		}
		if (obj instanceof String) {
			return obj;
		}else if (obj instanceof java.sql.Clob) {
			return clobToObject((java.sql.Clob) obj);
		}else if (obj instanceof byte[]) {
			return (byte[])obj;
		}else if (obj instanceof java.sql.Timestamp){
			return timestampToString((java.sql.Timestamp)obj);
		}
		return obj.toString();
	}
	private byte[] clobToObject(java.sql.Clob clob) {
		String s = "";
		if(clob != null){
			try {
				s = clob.getSubString(1, (int)clob.length());
			} catch (SQLException e) {
			}
		}
		return s.getBytes();
	}
	private String timestampToString(java.sql.Timestamp tm){
		if(tm != null){
			return dateFormat.format(tm);	
		}
		return "";
	}
	/*
	public List<String> lstColumns; // 字段列表
	public List<Object[]> lstDataList; // 返回的数据
	public int nNowPositon; // 当前位置
	private String[] arrResult; // 用于返回的数组，让不要重复分配
	DateFormat df = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
	

	// 设置完成
	protected void onSetComplete() {
		arrResult = new String[lstColumns.size()];
		nNowPositon = -1;
	}

	// 获得列头
	public String[] GetColumn() {
		return lstColumns.toArray(new String[lstColumns.size()]);
	}

	public int ColumnCount() {
		return lstColumns.size();
	}

	// 判断是否还有数据
	public boolean Next() {
		int nSize = lstDataList.size();
		if (nSize == 0) {
			return false;
		}
		nNowPositon++;
		if (nNowPositon >= nSize) {
			return false;
		}
		return true;
	}
	//获取一行的值
	public Object[] GetRow(){
		return null;
	}
	// 获取值
	public String GetValue(int nCol) {
		if (nNowPositon >= lstDataList.size()) {
			return "";
		}
		Object obj = null;
		try {
			obj = lstDataList.get(nNowPositon)[nCol];
		} catch (Exception e) {
			e.printStackTrace();
			return "";
		}
		if (obj == null) {
			return "";
		}
		if (obj instanceof java.sql.Clob) {
			return clobToString((java.sql.Clob) obj);
		}else if (obj instanceof byte[]) {
			return new String((byte[])obj);
		}else if (obj instanceof java.sql.Timestamp){
			return timestampToString((java.sql.Timestamp)obj);
		}
		return obj.toString();
	}

	// 关闭
	public void Close() {
		lstColumns = null;
		lstDataList = null;
		nNowPositon = -1;
		arrResult = null;
	}

	private String clobToString(java.sql.Clob clob) {
		String s = "";
		if(clob != null){
			try {
				s = clob.getSubString(1, (int)clob.length());
			} catch (SQLException e) {
				e.printStackTrace();
			}
		}
		return s;
	}
	private String timestampToString(java.sql.Timestamp tm){
		String s = "";
		if(tm != null){
			s = df.format(tm);	
		}
		return s;
	}*/
}