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
	B_price_ft_il    sql.NullFloat64

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
	C_price_ft_il    sql.NullFloat64

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
	P_price_ft_il    sql.NullFloat64

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
	V_price_ft_il    sql.NullFloat64
}

func GetPrice(db *sql.DB, mem_id string, errlog *log.Logger) BasePrice {
	price := BasePrice{}

	var priceSQL string

	priceSQL = `
	WITH RECURSIVE cte AS (
	  SELECT
	    mem_id AS _id,
	    mrg_recommend_mem_id
	  FROM
	    cb_member_register
	  WHERE
	    mem_id = $1
	  UNION ALL
	  SELECT
	    r.mem_id AS _id,
	    r.mrg_recommend_mem_id
	  FROM
	    cb_member_register r
	  INNER JOIN cte ON r.mrg_recommend_mem_id = cte._id
	)
	SELECT 
	  i.mad_price_at AS c_mad_price_at,
	  i.mad_price_ft AS c_mad_price_ft,
	  i.mad_price_ft_img AS c_mad_price_ft_img,
	  i.mad_price_ft_w_img AS c_mad_price_ft_w_img,
	  i.mad_price_grs AS c_mad_price_grs,
	  i.mad_price_nas AS c_mad_price_nas,
	  i.mad_price_grs_sms AS c_mad_price_grs_sms,
	  i.mad_price_nas_sms AS c_mad_price_nas_sms,
	  i.mad_price_015 AS c_mad_price_015,
	  i.mad_price_phn AS c_mad_price_phn,
	  i.mad_price_sms AS c_mad_price_sms,
	  i.mad_price_lms AS c_mad_price_lms,
	  i.mad_price_mms AS c_mad_price_mms,
	  i.mad_price_grs_mms AS c_mad_price_grs_mms,
	  i.mad_price_nas_mms AS c_mad_price_nas_mms,
	  i.mad_price_smt AS c_mad_price_smt,
	  i.mad_price_smt_sms AS c_mad_price_smt_sms,
	  i.mad_price_smt_mms AS c_mad_price_smt_mms,
	  i.mad_price_imc AS c_mad_price_imc,
	  i.mad_price_rcs AS c_mad_price_rcs,
	  i.mad_price_rcs_sms AS c_mad_price_rcs_sms,
	  i.mad_price_rcs_mms AS c_mad_price_rcs_mms,
	  i.mad_price_rcs_tem AS c_mad_price_rcs_tem,
	  i.mad_price_cs AS c_mad_price_cs,
	  i.mad_price_il AS c_mad_price_il,
	  a.mad_price_at AS p_mad_price_at,
	  a.mad_price_ft AS p_mad_price_ft,
	  a.mad_price_ft_img AS p_mad_price_ft_img,
	  a.mad_price_ft_w_img AS p_mad_price_ft_w_img,
	  a.mad_price_grs AS p_mad_price_grs,
	  a.mad_price_nas AS p_mad_price_nas,
	  a.mad_price_grs_sms AS p_mad_price_grs_sms,
	  a.mad_price_nas_sms AS p_mad_price_nas_sms,
	  a.mad_price_015 AS p_mad_price_015,
	  a.mad_price_phn AS p_mad_price_phn,
	  a.mad_price_sms AS p_mad_price_sms,
	  a.mad_price_lms AS p_mad_price_lms,
	  a.mad_price_mms AS p_mad_price_mms,
	  a.mad_price_grs_mms AS p_mad_price_grs_mms,
	  a.mad_price_nas_mms AS p_mad_price_nas_mms,
	  a.mad_price_smt AS p_mad_price_smt,
	  a.mad_price_smt_sms AS p_mad_price_smt_sms,
	  a.mad_price_smt_mms AS p_mad_price_smt_mms,
	  a.mad_price_imc AS p_mad_price_imc,
	  a.mad_price_rcs AS p_mad_price_rcs,
	  a.mad_price_rcs_sms AS p_mad_price_rcs_sms,
	  a.mad_price_rcs_mms AS p_mad_price_rcs_mms,
	  a.mad_price_rcs_tem AS p_mad_price_rcs_tem,
	  a.mad_price_cs AS p_mad_price_cs,
	  a.mad_price_il AS p_mad_price_il
	FROM 
	  cb_wt_member_addon i 
	  LEFT JOIN cb_wt_member_addon a ON 1=1 
	  INNER JOIN cb_member b ON a.mad_mem_id = b.mem_id 
	  INNER JOIN cte ON a.mad_mem_id = cte.mrg_recommend_mem_id 
	WHERE 
	  i.mad_mem_id = $2
	ORDER BY 
	  b.mem_level DESC
	`
	rows, err := db.Query(priceSQL, mem_id, mem_id)
	if err != nil {
		errlog.Println("Price 조회 중 오류 발생 sql : ", priceSQL)
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

	priceSQL = `select c.wst_price_at
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
		errlog.Println("Price 조회 중 오류 발생 sql : ", priceSQL)
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
		errlog.Println("Price 조회 중 오류 발생 sql : ", priceSQL)
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
