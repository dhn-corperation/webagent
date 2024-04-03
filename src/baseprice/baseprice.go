package baseprice

import (
	"database/sql"
	"log"
)

type BasePrice struct {
	B_price_ft       sql.NullFloat64
	B_price_ft_img   sql.NullFloat64
	B_price_ft_w_img sql.NullFloat64
	B_price_at       sql.NullFloat64
	B_price_sms      sql.NullFloat64
	B_price_lms      sql.NullFloat64
	B_price_mms      sql.NullFloat64
	B_price_phn      sql.NullFloat64
	B_price_015      sql.NullFloat64
	B_price_grs      sql.NullFloat64
	B_price_grs_sms  sql.NullFloat64
	B_price_nas      sql.NullFloat64
	B_price_nas_sms  sql.NullFloat64
	B_price_dooit    sql.NullFloat64
	B_price_grs_mms  sql.NullFloat64
	B_price_nas_mms  sql.NullFloat64
	B_price_smt      sql.NullFloat64
	B_price_smt_sms  sql.NullFloat64
	B_price_smt_mms  sql.NullFloat64
	B_price_imc      sql.NullFloat64
	B_price_rcs      sql.NullFloat64
	B_price_rcs_sms  sql.NullFloat64
	B_price_rcs_mms  sql.NullFloat64
	B_price_rcs_tem  sql.NullFloat64
	B_price_ft_cs    sql.NullFloat64
	B_price_ft_il	 sql.NullFloat64

	C_price_ft       sql.NullFloat64
	C_price_ft_img   sql.NullFloat64
	C_price_ft_w_img sql.NullFloat64
	C_price_at       sql.NullFloat64
	C_price_sms      sql.NullFloat64
	C_price_lms      sql.NullFloat64
	C_price_mms      sql.NullFloat64
	C_price_phn      sql.NullFloat64
	C_price_015      sql.NullFloat64
	C_price_grs      sql.NullFloat64
	C_price_grs_sms  sql.NullFloat64
	C_price_nas      sql.NullFloat64
	C_price_nas_sms  sql.NullFloat64
	C_price_dooit    sql.NullFloat64
	C_price_grs_mms  sql.NullFloat64
	C_price_nas_mms  sql.NullFloat64
	C_price_smt      sql.NullFloat64
	C_price_smt_sms  sql.NullFloat64
	C_price_smt_mms  sql.NullFloat64
	C_price_imc      sql.NullFloat64
	C_price_rcs      sql.NullFloat64
	C_price_rcs_sms  sql.NullFloat64
	C_price_rcs_mms  sql.NullFloat64
	C_price_rcs_tem  sql.NullFloat64
	C_price_ft_cs    sql.NullFloat64
	C_price_ft_il	 sql.NullFloat64

	P_price_ft       sql.NullFloat64
	P_price_ft_img   sql.NullFloat64
	P_price_ft_w_img sql.NullFloat64
	P_price_at       sql.NullFloat64
	P_price_sms      sql.NullFloat64
	P_price_lms      sql.NullFloat64
	P_price_mms      sql.NullFloat64
	P_price_phn      sql.NullFloat64
	P_price_015      sql.NullFloat64
	P_price_grs      sql.NullFloat64
	P_price_grs_sms  sql.NullFloat64
	P_price_nas      sql.NullFloat64
	P_price_nas_sms  sql.NullFloat64
	P_price_dooit    sql.NullFloat64
	P_price_grs_mms  sql.NullFloat64
	P_price_nas_mms  sql.NullFloat64
	P_price_smt      sql.NullFloat64
	P_price_smt_sms  sql.NullFloat64
	P_price_smt_mms  sql.NullFloat64
	P_price_imc      sql.NullFloat64
	P_price_rcs      sql.NullFloat64
	P_price_rcs_sms  sql.NullFloat64
	P_price_rcs_mms  sql.NullFloat64
	P_price_rcs_tem  sql.NullFloat64
	P_price_ft_cs    sql.NullFloat64
	P_price_ft_il	 sql.NullFloat64

	V_price_ft       sql.NullFloat64
	V_price_ft_img   sql.NullFloat64
	V_price_ft_w_img sql.NullFloat64
	V_price_at       sql.NullFloat64
	V_price_smt      sql.NullFloat64
	V_price_smt_sms  sql.NullFloat64
	V_price_smt_mms  sql.NullFloat64
	V_price_imc      sql.NullFloat64
	V_price_rcs      sql.NullFloat64
	V_price_rcs_sms  sql.NullFloat64
	V_price_rcs_mms  sql.NullFloat64
	V_price_rcs_tem  sql.NullFloat64
	V_price_ft_cs    sql.NullFloat64
	V_price_ft_il	 sql.NullFloat64
	
}

func GetPrice(db *sql.DB, mem_id string, errlog *log.Logger) BasePrice {
	price := BasePrice{}

	var priceSQL string
	priceSQL = "select SQL_NO_CACHE i.mad_price_at     as c_mad_price_at    " +
		"				,i.mad_price_ft     as c_mad_price_ft    " +
		"				,i.mad_price_ft_img as c_mad_price_ft_img" +
		"				,i.mad_price_ft_w_img as c_mad_price_ft_w_img" +
		"				,i.mad_price_grs    as c_mad_price_grs   " +
		"				,i.mad_price_nas    as c_mad_price_nas   " +
		"				,i.mad_price_grs_sms as c_mad_price_grs_sms   " +
		"				,i.mad_price_nas_sms as c_mad_price_nas_sms   " +
		"				,i.mad_price_015    as c_mad_price_015   " +
		"				,i.mad_price_phn    as c_mad_price_phn   " +
		"				,i.mad_price_sms    as c_mad_price_sms   " +
		"				,i.mad_price_lms    as c_mad_price_lms   " +
		"				,i.mad_price_mms    as c_mad_price_mms   " +
		"				,i.mad_price_grs_mms    as c_mad_price_grs_mms   " +
		"				,i.mad_price_nas_mms    as c_mad_price_nas_mms   " +
		"				,i.mad_price_smt    as c_mad_price_smt   " +
		"				,i.mad_price_smt_sms    as c_mad_price_smt_sms   " +
		"				,i.mad_price_smt_mms    as c_mad_price_smt_mms   " +
		"				,i.mad_price_imc    as c_mad_price_imc   " +
		"				,i.mad_price_rcs    as c_mad_price_rcs   " +
		"				,i.mad_price_rcs_sms    as c_mad_price_rcs_sms   " +
		"				,i.mad_price_rcs_mms    as c_mad_price_rcs_mms   " +
		"				,i.mad_price_rcs_tem    as c_mad_price_rcs_tem   " +
		"				,i.mad_price_cs    as c_mad_price_cs   " +
		"				,i.mad_price_il    as c_mad_price_il   " +
		"				,a.mad_price_at     as p_mad_price_at    " +
		"				,a.mad_price_ft     as p_mad_price_ft    " +
		"				,a.mad_price_ft_img as p_mad_price_ft_img" +
		"				,a.mad_price_ft_w_img as p_mad_price_ft_w_img" +
		"				,a.mad_price_grs    as p_mad_price_grs   " +
		"				,a.mad_price_nas    as p_mad_price_nas   " +
		"				,a.mad_price_grs_sms as p_mad_price_grs_sms   " +
		"				,a.mad_price_nas_sms as p_mad_price_nas_sms   " +
		"				,a.mad_price_015    as p_mad_price_015   " +
		"				,a.mad_price_phn    as p_mad_price_phn   " +
		"				,a.mad_price_sms    as p_mad_price_sms   " +
		"				,a.mad_price_lms    as p_mad_price_lms   " +
		"				,a.mad_price_mms    as p_mad_price_mms   " +
		"				,a.mad_price_grs_mms    as p_mad_price_grs_mms   " +
		"				,a.mad_price_nas_mms    as p_mad_price_nas_mms   " +
		"				,a.mad_price_smt    as p_mad_price_smt   " +
		"				,a.mad_price_smt_sms    as p_mad_price_smt_sms   " +
		"				,a.mad_price_smt_mms    as p_mad_price_smt_mms   " +
		"				,a.mad_price_imc    as p_mad_price_imc   " +
		"				,a.mad_price_rcs    as p_mad_price_rcs   " +
		"				,a.mad_price_rcs_sms    as p_mad_price_rcs_sms   " +
		"				,a.mad_price_rcs_mms    as p_mad_price_rcs_mms   " +
		"				,a.mad_price_rcs_tem    as p_mad_price_rcs_tem   " +
		"				,a.mad_price_cs    as p_mad_price_cs   " +
		"				,a.mad_price_il    as p_mad_price_il   " +
		"			from" +
		"				cb_wt_member_addon i left join" +
		"				cb_wt_member_addon a on 1=1 inner join" +
		"				cb_member b on a.mad_mem_id=b.mem_id inner join" +
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
			&price.C_price_grs,
			&price.C_price_nas,
			&price.C_price_grs_sms,
			&price.C_price_nas_sms,
			&price.C_price_015,
			&price.C_price_phn,
			&price.C_price_sms,
			&price.C_price_lms,
			&price.C_price_mms,
			&price.C_price_grs_mms,
			&price.C_price_nas_mms,
			&price.C_price_smt,
			&price.C_price_smt_sms,
			&price.C_price_smt_mms,
			&price.C_price_imc,
			&price.C_price_rcs,
			&price.C_price_rcs_sms,
			&price.C_price_rcs_mms,
			&price.C_price_rcs_tem,
			&price.C_price_ft_cs,
			&price.C_price_ft_il,
			&price.P_price_at,
			&price.P_price_ft,
			&price.P_price_ft_img,
			&price.P_price_ft_w_img,
			&price.P_price_grs,
			&price.P_price_nas,
			&price.P_price_grs_sms,
			&price.P_price_nas_sms,
			&price.P_price_015,
			&price.P_price_phn,
			&price.P_price_sms,
			&price.P_price_lms,
			&price.P_price_mms,
			&price.P_price_grs_mms,
			&price.P_price_nas_mms,
			&price.P_price_smt,
			&price.P_price_smt_sms,
			&price.P_price_smt_mms,
			&price.P_price_imc,
			&price.P_price_rcs,
			&price.P_price_rcs_sms,
			&price.P_price_rcs_mms,
			&price.P_price_rcs_tem,			
			&price.P_price_ft_cs,
			&price.P_price_ft_il,
			)

		if err != nil {
			errlog.Println(err)
		}

	}

	priceSQL = `select SQL_NO_CACHE c.wst_price_at
	,c.wst_price_ft
	,c.wst_price_ft_img
	,c.wst_price_ft_w_img
	,c.wst_price_grs
	,c.wst_price_nas
	,c.wst_price_grs_sms
	,c.wst_price_nas_sms
	,c.wst_price_015
	,c.wst_price_phn
	,c.wst_price_sms
	,c.wst_price_lms
	,c.wst_price_mms
	,c.wst_price_grs_mms
	,c.wst_price_nas_mms
	,c.wst_price_smt
	,c.wst_price_smt_sms
	,c.wst_price_smt_mms
	,c.wst_price_imc
	, wst_price_rcs
	, wst_price_rcs_sms
	, wst_price_rcs_mms
	, wst_price_rcs_tem
	, wst_price_cs
	, wst_price_il 
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
			&price.B_price_grs,
			&price.B_price_nas,
			&price.B_price_grs_sms,
			&price.B_price_nas_sms,
			&price.B_price_015,
			&price.B_price_phn,
			&price.B_price_sms,
			&price.B_price_lms,
			&price.B_price_mms,
			&price.B_price_grs_mms,
			&price.B_price_nas_mms,
			&price.B_price_smt,
			&price.B_price_smt_sms,
			&price.B_price_smt_mms,
			&price.B_price_imc,
			&price.B_price_rcs,
			&price.B_price_rcs_sms,
			&price.B_price_rcs_mms,
			&price.B_price_rcs_tem,
			&price.B_price_ft_cs,
			&price.B_price_ft_il,
		)
		if err != nil {
			errlog.Println(err)
		}
	}

	priceSQL = `select vad_price_ft
	, vad_price_ft_img
	, vad_price_at
	, vad_price_smt
	, vad_price_smt_sms
	, vad_price_smt_mms
	, vad_price_imc
	, vad_price_rcs
	, vad_price_rcs_sms
	, vad_price_rcs_mms
	, vad_price_rcs_tem 
	, vad_price_cs
	, vad_price_il
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
			&price.V_price_imc,
			&price.V_price_rcs,
			&price.V_price_rcs_sms,
			&price.V_price_rcs_mms,
			&price.V_price_rcs_tem,
			&price.V_price_ft_cs,
			&price.V_price_ft_il,
		)
		if err != nil {
			errlog.Println(err)
		}
	}

	return price
}
