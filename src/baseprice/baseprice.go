package baseprice

import (
	"log"
	"database/sql"
)

type BasePrice struct {
	B_price_ft       sql.NullFloat64
	B_price_ft_img   sql.NullFloat64
	B_price_ft_w_img sql.NullFloat64
	B_price_at       sql.NullFloat64
	B_price_smt      sql.NullFloat64
	B_price_smt_sms  sql.NullFloat64
	B_price_smt_mms  sql.NullFloat64
	B_price_rcs      sql.NullFloat64
	B_price_rcs_sms  sql.NullFloat64
	B_price_rcs_mms  sql.NullFloat64
	B_price_rcs_tem  sql.NullFloat64
	B_price_ft_cs    sql.NullFloat64
	B_price_ft_il	 sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
	B_price_bm_t_m 	 sql.NullFloat64
	B_price_bm_t_n 	 sql.NullFloat64
	B_price_bm_t_i 	 sql.NullFloat64
	B_price_bm_t_f 	 sql.NullFloat64
	B_price_bm_b1    sql.NullFloat64
	B_price_bm_b2    sql.NullFloat64
	B_price_bm_b3    sql.NullFloat64
	B_price_bm_b4    sql.NullFloat64
	B_price_bm_b5    sql.NullFloat64
	B_price_bm_b6    sql.NullFloat64
	B_price_bm_b7    sql.NullFloat64
	B_price_bm_b8    sql.NullFloat64
	B_price_bm_c1    sql.NullFloat64
	B_price_bm_c2    sql.NullFloat64
	B_price_bm_c3    sql.NullFloat64
	B_price_bm_c4    sql.NullFloat64
	B_price_bm_c5    sql.NullFloat64
	B_price_bm_c6    sql.NullFloat64
	B_price_bm_c7    sql.NullFloat64
	B_price_bm_c8    sql.NullFloat64
	B_price_bm_d1    sql.NullFloat64
	B_price_bm_d2    sql.NullFloat64
	B_price_bm_d3    sql.NullFloat64
	B_price_bm_d4    sql.NullFloat64
	B_price_bm_d5    sql.NullFloat64
	B_price_bm_d6    sql.NullFloat64
	B_price_bm_d7    sql.NullFloat64
	B_price_bm_d8    sql.NullFloat64
	B_price_bm_f     sql.NullFloat64
	B_price_bm_nf    sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////


	C_price_ft       sql.NullFloat64
	C_price_ft_img   sql.NullFloat64
	C_price_ft_w_img sql.NullFloat64
	C_price_at       sql.NullFloat64
	C_price_smt      sql.NullFloat64
	C_price_smt_sms  sql.NullFloat64
	C_price_smt_mms  sql.NullFloat64
	C_price_rcs      sql.NullFloat64
	C_price_rcs_sms  sql.NullFloat64
	C_price_rcs_mms  sql.NullFloat64
	C_price_rcs_tem  sql.NullFloat64
	C_price_ft_cs    sql.NullFloat64
	C_price_ft_il	 sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
	C_price_bm_t_m   sql.NullFloat64
	C_price_bm_t_n   sql.NullFloat64
	C_price_bm_t_i   sql.NullFloat64
	C_price_bm_t_f   sql.NullFloat64
	C_price_bm_b1    sql.NullFloat64
	C_price_bm_b2    sql.NullFloat64
	C_price_bm_b3    sql.NullFloat64
	C_price_bm_b4    sql.NullFloat64
	C_price_bm_b5    sql.NullFloat64
	C_price_bm_b6    sql.NullFloat64
	C_price_bm_b7    sql.NullFloat64
	C_price_bm_b8    sql.NullFloat64
	C_price_bm_c1    sql.NullFloat64
	C_price_bm_c2    sql.NullFloat64
	C_price_bm_c3    sql.NullFloat64
	C_price_bm_c4    sql.NullFloat64
	C_price_bm_c5    sql.NullFloat64
	C_price_bm_c6    sql.NullFloat64
	C_price_bm_c7    sql.NullFloat64
	C_price_bm_c8    sql.NullFloat64
	C_price_bm_d1    sql.NullFloat64
	C_price_bm_d2    sql.NullFloat64
	C_price_bm_d3    sql.NullFloat64
	C_price_bm_d4    sql.NullFloat64
	C_price_bm_d5    sql.NullFloat64
	C_price_bm_d6    sql.NullFloat64
	C_price_bm_d7    sql.NullFloat64
	C_price_bm_d8    sql.NullFloat64
	C_price_bm_f     sql.NullFloat64
	C_price_bm_nf    sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 

	P_price_ft       sql.NullFloat64
	P_price_ft_img   sql.NullFloat64
	P_price_ft_w_img sql.NullFloat64
	P_price_at       sql.NullFloat64
	P_price_smt      sql.NullFloat64
	P_price_smt_sms  sql.NullFloat64
	P_price_smt_mms  sql.NullFloat64
	P_price_rcs      sql.NullFloat64
	P_price_rcs_sms  sql.NullFloat64
	P_price_rcs_mms  sql.NullFloat64
	P_price_rcs_tem  sql.NullFloat64
	P_price_ft_cs    sql.NullFloat64
	P_price_ft_il	 sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
	P_price_bm_t_m   sql.NullFloat64
	P_price_bm_t_n   sql.NullFloat64
	P_price_bm_t_i   sql.NullFloat64
	P_price_bm_t_f   sql.NullFloat64
	P_price_bm_b1    sql.NullFloat64
	P_price_bm_b2    sql.NullFloat64
	P_price_bm_b3    sql.NullFloat64
	P_price_bm_b4    sql.NullFloat64
	P_price_bm_b5    sql.NullFloat64
	P_price_bm_b6    sql.NullFloat64
	P_price_bm_b7    sql.NullFloat64
	P_price_bm_b8    sql.NullFloat64
	P_price_bm_c1    sql.NullFloat64
	P_price_bm_c2    sql.NullFloat64
	P_price_bm_c3    sql.NullFloat64
	P_price_bm_c4    sql.NullFloat64
	P_price_bm_c5    sql.NullFloat64
	P_price_bm_c6    sql.NullFloat64
	P_price_bm_c7    sql.NullFloat64
	P_price_bm_c8    sql.NullFloat64
	P_price_bm_d1    sql.NullFloat64
	P_price_bm_d2    sql.NullFloat64
	P_price_bm_d3    sql.NullFloat64
	P_price_bm_d4    sql.NullFloat64
	P_price_bm_d5    sql.NullFloat64
	P_price_bm_d6    sql.NullFloat64
	P_price_bm_d7    sql.NullFloat64
	P_price_bm_d8    sql.NullFloat64
	P_price_bm_f     sql.NullFloat64
	P_price_bm_nf    sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 

	V_price_ft       sql.NullFloat64
	V_price_ft_img   sql.NullFloat64
	V_price_ft_w_img sql.NullFloat64
	V_price_at       sql.NullFloat64
	V_price_smt      sql.NullFloat64
	V_price_smt_sms  sql.NullFloat64
	V_price_smt_mms  sql.NullFloat64
	V_price_rcs      sql.NullFloat64
	V_price_rcs_sms  sql.NullFloat64
	V_price_rcs_mms  sql.NullFloat64
	V_price_rcs_tem  sql.NullFloat64
	V_price_ft_cs    sql.NullFloat64
	V_price_ft_il	 sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
	V_price_bm_t_m   sql.NullFloat64
	V_price_bm_t_n   sql.NullFloat64
	V_price_bm_t_i   sql.NullFloat64
	V_price_bm_t_f   sql.NullFloat64
	V_price_bm_b1    sql.NullFloat64
	V_price_bm_b2    sql.NullFloat64
	V_price_bm_b3    sql.NullFloat64
	V_price_bm_b4    sql.NullFloat64
	V_price_bm_b5    sql.NullFloat64
	V_price_bm_b6    sql.NullFloat64
	V_price_bm_b7    sql.NullFloat64
	V_price_bm_b8    sql.NullFloat64
	V_price_bm_c1    sql.NullFloat64
	V_price_bm_c2    sql.NullFloat64
	V_price_bm_c3    sql.NullFloat64
	V_price_bm_c4    sql.NullFloat64
	V_price_bm_c5    sql.NullFloat64
	V_price_bm_c6    sql.NullFloat64
	V_price_bm_c7    sql.NullFloat64
	V_price_bm_c8    sql.NullFloat64
	V_price_bm_d1    sql.NullFloat64
	V_price_bm_d2    sql.NullFloat64
	V_price_bm_d3    sql.NullFloat64
	V_price_bm_d4    sql.NullFloat64
	V_price_bm_d5    sql.NullFloat64
	V_price_bm_d6    sql.NullFloat64
	V_price_bm_d7    sql.NullFloat64
	V_price_bm_d8    sql.NullFloat64
	V_price_bm_f     sql.NullFloat64
	V_price_bm_nf    sql.NullFloat64
	//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
	
}

func GetPrice(db *sql.DB, mem_id string, errlog *log.Logger) BasePrice {
	price := BasePrice{}

	var priceSQL string
	priceSQL = "select SQL_NO_CACHE i.mad_price_at     as c_mad_price_at    " +
		"				,i.mad_price_ft     as c_mad_price_ft    " +
		"				,i.mad_price_ft_img as c_mad_price_ft_img" +
		"				,i.mad_price_ft_w_img as c_mad_price_ft_w_img" +
		"				,i.mad_price_smt    as c_mad_price_smt   " +
		"				,i.mad_price_smt_sms    as c_mad_price_smt_sms   " +
		"				,i.mad_price_smt_mms    as c_mad_price_smt_mms   " +
		"				,i.mad_price_rcs    as c_mad_price_rcs   " +
		"				,i.mad_price_rcs_sms    as c_mad_price_rcs_sms   " +
		"				,i.mad_price_rcs_mms    as c_mad_price_rcs_mms   " +
		"				,i.mad_price_rcs_tem    as c_mad_price_rcs_tem   " +
		"				,i.mad_price_cs    as c_mad_price_cs   " +
		"				,i.mad_price_il    as c_mad_price_il   " +
//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
		"				,i.mad_price_bm_t_m    as c_mad_price_bm_t_m   " +
		"				,i.mad_price_bm_t_n    as c_mad_price_bm_t_n   " +
		"				,i.mad_price_bm_t_i    as c_mad_price_bm_t_i   " +
		"				,i.mad_price_bm_t_f    as c_mad_price_bm_t_f   " +
		"				,i.mad_price_bm_b1    as c_mad_price_bm_b1   " +
		"				,i.mad_price_bm_b2    as c_mad_price_bm_b2   " +
		"				,i.mad_price_bm_b3    as c_mad_price_bm_b3   " +
		"				,i.mad_price_bm_b4    as c_mad_price_bm_b4   " +
		"				,i.mad_price_bm_b5    as c_mad_price_bm_b5   " +
		"				,i.mad_price_bm_b6    as c_mad_price_bm_b6   " +
		"				,i.mad_price_bm_b7    as c_mad_price_bm_b7   " +
		"				,i.mad_price_bm_b8    as c_mad_price_bm_b8   " +
		"				,i.mad_price_bm_c1    as c_mad_price_bm_c1   " +
		"				,i.mad_price_bm_c2    as c_mad_price_bm_c2   " +
		"				,i.mad_price_bm_c3    as c_mad_price_bm_c3   " +
		"				,i.mad_price_bm_c4    as c_mad_price_bm_c4   " +
		"				,i.mad_price_bm_c5    as c_mad_price_bm_c5   " +
		"				,i.mad_price_bm_c6    as c_mad_price_bm_c6   " +
		"				,i.mad_price_bm_c7    as c_mad_price_bm_c7   " +
		"				,i.mad_price_bm_c8    as c_mad_price_bm_c8   " +
		"				,i.mad_price_bm_d1    as c_mad_price_bm_d1   " +
		"				,i.mad_price_bm_d2    as c_mad_price_bm_d2   " +
		"				,i.mad_price_bm_d3    as c_mad_price_bm_d3   " +
		"				,i.mad_price_bm_d4    as c_mad_price_bm_d4   " +
		"				,i.mad_price_bm_d5    as c_mad_price_bm_d5   " +
		"				,i.mad_price_bm_d6    as c_mad_price_bm_d6   " +
		"				,i.mad_price_bm_d7    as c_mad_price_bm_d7   " +
		"				,i.mad_price_bm_d8    as c_mad_price_bm_d8   " +
		"				,i.mad_price_bm_f     as c_mad_price_bm_f    " +
		"				,i.mad_price_bm_nf    as c_mad_price_bm_nf   " +
//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
		"				,a.mad_price_at     as p_mad_price_at    " +
		"				,a.mad_price_ft     as p_mad_price_ft    " +
		"				,a.mad_price_ft_img as p_mad_price_ft_img" +
		"				,a.mad_price_ft_w_img as p_mad_price_ft_w_img" +
		"				,a.mad_price_smt    as p_mad_price_smt   " +
		"				,a.mad_price_smt_sms    as p_mad_price_smt_sms   " +
		"				,a.mad_price_smt_mms    as p_mad_price_smt_mms   " +
		"				,a.mad_price_rcs    as p_mad_price_rcs   " +
		"				,a.mad_price_rcs_sms    as p_mad_price_rcs_sms   " +
		"				,a.mad_price_rcs_mms    as p_mad_price_rcs_mms   " +
		"				,a.mad_price_rcs_tem    as p_mad_price_rcs_tem   " +
		"				,a.mad_price_cs    as p_mad_price_cs   " +
		"				,a.mad_price_il    as p_mad_price_il   " +
//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
		"				,a.mad_price_bm_t_m    as p_mad_price_bm_t_m   " +
		"				,a.mad_price_bm_t_n    as p_mad_price_bm_t_n   " +
		"				,a.mad_price_bm_t_i    as p_mad_price_bm_t_i   " +
		"				,a.mad_price_bm_t_f    as p_mad_price_bm_t_f   " +
		"				,a.mad_price_bm_b1    as p_mad_price_bm_b1   " +
		"				,a.mad_price_bm_b2    as p_mad_price_bm_b2   " +
		"				,a.mad_price_bm_b3    as p_mad_price_bm_b3   " +
		"				,a.mad_price_bm_b4    as p_mad_price_bm_b4   " +
		"				,a.mad_price_bm_b5    as p_mad_price_bm_b5   " +
		"				,a.mad_price_bm_b6    as p_mad_price_bm_b6   " +
		"				,a.mad_price_bm_b7    as p_mad_price_bm_b7   " +
		"				,a.mad_price_bm_b8    as p_mad_price_bm_b8   " +
		"				,a.mad_price_bm_c1    as p_mad_price_bm_c1   " +
		"				,a.mad_price_bm_c2    as p_mad_price_bm_c2   " +
		"				,a.mad_price_bm_c3    as p_mad_price_bm_c3   " +
		"				,a.mad_price_bm_c4    as p_mad_price_bm_c4   " +
		"				,a.mad_price_bm_c5    as p_mad_price_bm_c5   " +
		"				,a.mad_price_bm_c6    as p_mad_price_bm_c6   " +
		"				,a.mad_price_bm_c7    as p_mad_price_bm_c7   " +
		"				,a.mad_price_bm_c8    as p_mad_price_bm_c8   " +
		"				,a.mad_price_bm_d1    as p_mad_price_bm_d1   " +
		"				,a.mad_price_bm_d2    as p_mad_price_bm_d2   " +
		"				,a.mad_price_bm_d3    as p_mad_price_bm_d3   " +
		"				,a.mad_price_bm_d4    as p_mad_price_bm_d4   " +
		"				,a.mad_price_bm_d5    as p_mad_price_bm_d5   " +
		"				,a.mad_price_bm_d6    as p_mad_price_bm_d6   " +
		"				,a.mad_price_bm_d7    as p_mad_price_bm_d7   " +
		"				,a.mad_price_bm_d8    as p_mad_price_bm_d8   " +
		"				,a.mad_price_bm_f     as p_mad_price_bm_f    " +
		"				,a.mad_price_bm_nf    as p_mad_price_bm_nf   " +
//////////////////////////////////////////////////// BM AREA //////////////////////////////////////////////////// 
		"			from" +
		"				cb_wt_member_addon i left join" +
		"				cb_wt_member_addon a on 1=1 inner join" +
		"				cb_member b on a.mad_mem_id=b.mem_id and b.mem_level <> 151 inner join" +
		"				(" +
		"					SELECT distinct @r AS _id, (SELECT  @r := mrg_recommend_mem_id FROM cb_member_register WHERE mem_id = _id ) AS mrg_recommend_mem_id" +
		"					FROM" +
		"						(SELECT  @r := " + mem_id + ", @cl := 0) vars,	cb_member_register h" +
		"					WHERE    @r <> 0" +
		"				) c on a.mad_mem_id=c.mrg_recommend_mem_id" +
		"			where i.mad_mem_id=" + mem_id +
		"			order by b.mem_level desc"
	//stdlog.Println(priceSQL)
	rows, err := db.Query(priceSQL)

	if err != nil {
		errlog.Println("Price 조회 중 오류 발생")
		errlog.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&price.C_price_at,
			&price.C_price_ft,
			&price.C_price_ft_img,
			&price.C_price_ft_w_img,
			&price.C_price_smt,
			&price.C_price_smt_sms,
			&price.C_price_smt_mms,
			&price.C_price_rcs,
			&price.C_price_rcs_sms,
			&price.C_price_rcs_mms,
			&price.C_price_rcs_tem,
			&price.C_price_ft_cs,
			&price.C_price_ft_il,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
			&price.C_price_bm_t_m,
			&price.C_price_bm_t_n,
			&price.C_price_bm_t_i,
			&price.C_price_bm_t_f,
			&price.C_price_bm_b1,
			&price.C_price_bm_b2,
			&price.C_price_bm_b3,
			&price.C_price_bm_b4,
			&price.C_price_bm_b5,
			&price.C_price_bm_b6,
			&price.C_price_bm_b7,
			&price.C_price_bm_b8,
			&price.C_price_bm_c1,
			&price.C_price_bm_c2,
			&price.C_price_bm_c3,
			&price.C_price_bm_c4,
			&price.C_price_bm_c5,
			&price.C_price_bm_c6,
			&price.C_price_bm_c7,
			&price.C_price_bm_c8,
			&price.C_price_bm_d1,
			&price.C_price_bm_d2,
			&price.C_price_bm_d3,
			&price.C_price_bm_d4,
			&price.C_price_bm_d5,
			&price.C_price_bm_d6,
			&price.C_price_bm_d7,
			&price.C_price_bm_d8,
			&price.C_price_bm_f,
			&price.C_price_bm_nf,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
			&price.P_price_at,
			&price.P_price_ft,
			&price.P_price_ft_img,
			&price.P_price_ft_w_img,
			&price.P_price_smt,
			&price.P_price_smt_sms,
			&price.P_price_smt_mms,
			&price.P_price_rcs,
			&price.P_price_rcs_sms,
			&price.P_price_rcs_mms,
			&price.P_price_rcs_tem,			
			&price.P_price_ft_cs,
			&price.P_price_ft_il,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
			&price.P_price_bm_t_m,
			&price.P_price_bm_t_n,
			&price.P_price_bm_t_i,
			&price.P_price_bm_t_f,
			&price.P_price_bm_b1,
			&price.P_price_bm_b2,
			&price.P_price_bm_b3,
			&price.P_price_bm_b4,
			&price.P_price_bm_b5,
			&price.P_price_bm_b6,
			&price.P_price_bm_b7,
			&price.P_price_bm_b8,
			&price.P_price_bm_c1,
			&price.P_price_bm_c2,
			&price.P_price_bm_c3,
			&price.P_price_bm_c4,
			&price.P_price_bm_c5,
			&price.P_price_bm_c6,
			&price.P_price_bm_c7,
			&price.P_price_bm_c8,
			&price.P_price_bm_d1,
			&price.P_price_bm_d2,
			&price.P_price_bm_d3,
			&price.P_price_bm_d4,
			&price.P_price_bm_d5,
			&price.P_price_bm_d6,
			&price.P_price_bm_d7,
			&price.P_price_bm_d8,
			&price.P_price_bm_f,
			&price.P_price_bm_nf,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
			)

		if err != nil {
			errlog.Println(err)
		}

	}

	priceSQL = `select SQL_NO_CACHE 
	 c.wst_price_at
	,c.wst_price_ft
	,c.wst_price_ft_img
	,c.wst_price_ft_w_img
	,c.wst_price_smt
	,c.wst_price_smt_sms
	,c.wst_price_smt_mms
	, wst_price_rcs
	, wst_price_rcs_sms
	, wst_price_rcs_mms
	, wst_price_rcs_tem
	, wst_price_cs
	, wst_price_il 
	, wst_price_bm_t_m
	, wst_price_bm_t_n
	, wst_price_bm_t_i
	, wst_price_bm_t_f
	, wst_price_bm_b1
	, wst_price_bm_b2
	, wst_price_bm_b3
	, wst_price_bm_b4
	, wst_price_bm_b5
	, wst_price_bm_b6
	, wst_price_bm_b7
	, wst_price_bm_b8
	, wst_price_bm_c1
	, wst_price_bm_c2
	, wst_price_bm_c3
	, wst_price_bm_c4
	, wst_price_bm_c5
	, wst_price_bm_c6
	, wst_price_bm_c7
	, wst_price_bm_c8
	, wst_price_bm_d1
	, wst_price_bm_d2
	, wst_price_bm_d3
	, wst_price_bm_d4
	, wst_price_bm_d5
	, wst_price_bm_d6
	, wst_price_bm_d7
	, wst_price_bm_d8
	, wst_price_bm_f
	, wst_price_bm_nf
	from cb_wt_setting c limit 1`


	rows, err = db.Query(priceSQL)

	if err != nil {
		errlog.Println("Price 조회 중 오류 발생")
		errlog.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&price.B_price_at,
			&price.B_price_ft,
			&price.B_price_ft_img,
			&price.B_price_ft_w_img,
			&price.B_price_smt,
			&price.B_price_smt_sms,
			&price.B_price_smt_mms,
			&price.B_price_rcs,
			&price.B_price_rcs_sms,
			&price.B_price_rcs_mms,
			&price.B_price_rcs_tem,
			&price.B_price_ft_cs,
			&price.B_price_ft_il,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
			&price.B_price_bm_t_m,
			&price.B_price_bm_t_n,
			&price.B_price_bm_t_i,
			&price.B_price_bm_t_f,
			&price.B_price_bm_b1,
			&price.B_price_bm_b2,
			&price.B_price_bm_b3,
			&price.B_price_bm_b4,
			&price.B_price_bm_b5,
			&price.B_price_bm_b6,
			&price.B_price_bm_b7,
			&price.B_price_bm_b8,
			&price.B_price_bm_c1,
			&price.B_price_bm_c2,
			&price.B_price_bm_c3,
			&price.B_price_bm_c4,
			&price.B_price_bm_c5,
			&price.B_price_bm_c6,
			&price.B_price_bm_c7,
			&price.B_price_bm_c8,
			&price.B_price_bm_d1,
			&price.B_price_bm_d2,
			&price.B_price_bm_d3,
			&price.B_price_bm_d4,
			&price.B_price_bm_d5,
			&price.B_price_bm_d6,
			&price.B_price_bm_d7,
			&price.B_price_bm_d8,
			&price.B_price_bm_f,
			&price.B_price_bm_nf,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
		)
		if err != nil {
			errlog.Println(err)
		}
	}

	priceSQL = `select
	  vad_price_ft
	, vad_price_ft_img
	, vad_price_at
	, vad_price_smt
	, vad_price_smt_sms
	, vad_price_smt_mms
	, vad_price_rcs
	, vad_price_rcs_sms
	, vad_price_rcs_mms
	, vad_price_rcs_tem 
	, vad_price_cs
	, vad_price_il
	, vad_price_bm_t_m
	, vad_price_bm_t_n
	, vad_price_bm_t_i
	, vad_price_bm_t_f
	, vad_price_bm_b1
	, vad_price_bm_b2
	, vad_price_bm_b3
	, vad_price_bm_b4
	, vad_price_bm_b5
	, vad_price_bm_b6
	, vad_price_bm_b7
	, vad_price_bm_b8
	, vad_price_bm_c1
	, vad_price_bm_c2
	, vad_price_bm_c3
	, vad_price_bm_c4
	, vad_price_bm_c5
	, vad_price_bm_c6
	, vad_price_bm_c7
	, vad_price_bm_c8
	, vad_price_bm_d1
	, vad_price_bm_d2
	, vad_price_bm_d3
	, vad_price_bm_d4
	, vad_price_bm_d5
	, vad_price_bm_d6
	, vad_price_bm_d7
	, vad_price_bm_d8
	, vad_price_bm_f
	, vad_price_bm_nf
	from cb_wt_voucher_addon where vad_mem_id = '` + mem_id + "'"

	rows, err = db.Query(priceSQL)

	if err != nil {
		errlog.Println("Price 조회 중 오류 발생")
		errlog.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&price.V_price_ft,
			&price.V_price_ft_img,
			&price.V_price_at,
			&price.V_price_smt,
			&price.V_price_smt_sms,
			&price.V_price_smt_mms,
			&price.V_price_rcs,
			&price.V_price_rcs_sms,
			&price.V_price_rcs_mms,
			&price.V_price_rcs_tem,
			&price.V_price_ft_cs,
			&price.V_price_ft_il,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
			&price.V_price_bm_t_m,
			&price.V_price_bm_t_n,
			&price.V_price_bm_t_i,
			&price.V_price_bm_t_f,
			&price.V_price_bm_b1,
			&price.V_price_bm_b2,
			&price.V_price_bm_b3,
			&price.V_price_bm_b4,
			&price.V_price_bm_b5,
			&price.V_price_bm_b6,
			&price.V_price_bm_b7,
			&price.V_price_bm_b8,
			&price.V_price_bm_c1,
			&price.V_price_bm_c2,
			&price.V_price_bm_c3,
			&price.V_price_bm_c4,
			&price.V_price_bm_c5,
			&price.V_price_bm_c6,
			&price.V_price_bm_c7,
			&price.V_price_bm_c8,
			&price.V_price_bm_d1,
			&price.V_price_bm_d2,
			&price.V_price_bm_d3,
			&price.V_price_bm_d4,
			&price.V_price_bm_d5,
			&price.V_price_bm_d6,
			&price.V_price_bm_d7,
			&price.V_price_bm_d8,
			&price.V_price_bm_f,
			&price.V_price_bm_nf,
//////////////////////////////////////////////////// BM AREA ////////////////////////////////////////////////////
		)
		if err != nil {
			errlog.Println(err)
		}
	}

	return price
}
