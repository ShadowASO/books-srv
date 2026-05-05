/*
---------------------------------------------------------------------------------------
File: mstypes.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package mstypes

// Tipo para construir objetos JSON
// Exemplo:
//
//	query := JsonMap{
//	    "size": 10,
//	    "query": JsonMap{
//	        "terms": JsonMap{
//	            "id_ctxt": []int{123, 456, 789},
//	        },
//	    },
//	}
type JsonMap map[string]interface{}
