terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:3301"
  access_token = "<SIGNOZ-API-KEY>"
}

resource "signoz_dashboard" "new_dashboard" {
  collapsable_rows_migrated = true
  description               = "test1"
  layout = jsonencode([
    {
      "h" : 8,
      "i" : "c5f29b09-8a63-44ba-825f-db91a3c79a54",
      "moved" : false,
      "static" : false,
      "w" : 6,
      "x" : 0,
      "y" : 0
    },
  ])
  name = "test1"
  panel_map = jsonencode(
    {
      "94c3ba3e-b5de-49da-8a4d-f4572585d2e6" : {
        "collapsed" : false,
        "widgets" : [
          {
            "h" : 6,
            "i" : "3d74094e-241b-4560-9128-abb1af79ae3c",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 172
          },
          {
            "h" : 6,
            "i" : "9a8ac524-afe4-457a-b17a-45f11c4f0fcf",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 6,
            "y" : 172
          },
          {
            "h" : 6,
            "i" : "f42a3a42-8099-4be9-a6b5-1528d1f1bdfa",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 178
          },
          {
            "h" : 6,
            "i" : "e796af4a-abe8-4ff4-9bba-0b7eb81c022a",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 6,
            "y" : 178
          },
          {
            "h" : 6,
            "i" : "9abc2374-d1ef-4ad6-958b-2355addcb245",
            "moved" : false,
            "static" : false,
            "w" : 12,
            "x" : 0,
            "y" : 184
          }
        ]
      },
      "a3b0f2cb-02de-41ef-b843-008e40b8f6d9" : {
        "collapsed" : false,
        "widgets" : [
          {
            "h" : 6,
            "i" : "92dd7aae-95eb-48ae-9d3a-6e062d468dab",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 57
          },
          {
            "h" : 6,
            "i" : "f984a994-50b8-4c1e-83b4-9a10ca128657",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 6,
            "y" : 57
          }
        ]
      },
      "cdae1cea-ab16-46db-9100-af2fb7971198" : {
        "collapsed" : false,
        "widgets" : [
          {
            "h" : 6,
            "i" : "a9013487-0479-4a09-a613-b3fc20e4d666",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 191
          },
          {
            "h" : 6,
            "i" : "3f186d43-6b86-4b71-962b-4ccddd8d7481",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 6,
            "y" : 191
          },
          {
            "h" : 6,
            "i" : "e2683847-faa3-42f6-bb3c-4367a3e76294",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 197
          },
          {
            "h" : 6,
            "i" : "64f1ea50-0c34-4e43-9912-0d9547e2e436",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 6,
            "y" : 197
          }
        ]
      },
      "f8abf828-e45d-4712-bb23-f6aec69bc4fa" : {
        "collapsed" : false,
        "widgets" : [
          {
            "h" : 3,
            "i" : "b33f0bad-e623-4c2b-b854-12270f211690",
            "moved" : false,
            "static" : false,
            "w" : 3,
            "x" : 0,
            "y" : 1
          },
          {
            "h" : 3,
            "i" : "5d89be47-9c43-4b0f-96c0-1dc72dbfa356",
            "moved" : false,
            "static" : false,
            "w" : 3,
            "x" : 3,
            "y" : 1
          },
          {
            "h" : 3,
            "i" : "0ce16128-ff8d-479d-8db2-10ac8fb47bc2",
            "moved" : false,
            "static" : false,
            "w" : 3,
            "x" : 6,
            "y" : 1
          },
          {
            "h" : 3,
            "i" : "f9df830b-cb02-4e70-b72a-f89a2e4f8196",
            "moved" : false,
            "static" : false,
            "w" : 3,
            "x" : 9,
            "y" : 1
          }
        ]
      },
      "f9d3d624-01e9-430c-b4a8-a1371fe35628" : {
        "collapsed" : false,
        "widgets" : [
          {
            "h" : 6,
            "i" : "5f43aa0f-cc5f-4a29-9cd7-dd0505db35a9",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 8
          },
          {
            "h" : 6,
            "i" : "ea1d1541-7787-4e09-8e5b-69be7f6af1ce",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 6,
            "y" : 8
          },
          {
            "h" : 6,
            "i" : "d497c2d7-372e-4670-9917-9bcbc25d0487",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 14
          },
          {
            "h" : 6,
            "i" : "45a54f00-081e-42b2-8134-c3848f4ac19e",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 6,
            "y" : 14
          }
        ]
      },
      "fbaade42-9fdb-4023-9a6b-b2a7000ab24c" : {
        "collapsed" : false,
        "widgets" : [
          {
            "h" : 6,
            "i" : "24e4d4c3-4ada-4e9d-bc3b-e19d314daa27",
            "moved" : false,
            "static" : false,
            "w" : 6,
            "x" : 0,
            "y" : 33
          }
        ]
      }
    }
  )
  tags             = ["node", "kubelet"]
  title            = "test1"
  uploaded_grafana = false
  variables = jsonencode(
    {
      "8ccae12a-9f04-4d27-8940-374446050530" : {
        "customValue" : "",
        "description" : "The k8s node name",
        "id" : "8ccae12a-9f04-4d27-8940-374446050530",
        "modificationUUID" : "4f18b48e-0d06-4f63-88fb-2299ca8a7b1f",
        "multiSelect" : true,
        "name" : "k8s_node_name",
        "order" : 1,
        "queryValue" : "SELECT JSONExtractString(labels, 'k8s_node_name') AS k8s_node_name\nFROM signoz_metrics.distributed_time_series_v4_1day\nWHERE metric_name = 'k8s_node_cpu_time' AND JSONExtractString(labels, 'k8s_cluster_name') = {{.k8s_cluster_name}}\nGROUP BY k8s_node_name",
        "showALLOption" : true,
        "sort" : "ASC",
        "textboxValue" : "",
        "type" : "QUERY",
        "selectedValue" : ["default"],
        "allSelected" : false
      },
      "e4232384-ad99-4479-8cec-e027a18921c4" : {
        "customValue" : "",
        "description" : "The k8s cluster name",
        "id" : "e4232384-ad99-4479-8cec-e027a18921c4",
        "key" : "e4232384-ad99-4479-8cec-e027a18921c4",
        "modificationUUID" : "c24c99c3-2390-4b53-860b-24dce4a7eab1",
        "multiSelect" : false,
        "name" : "k8s_cluster_name",
        "order" : 0,
        "queryValue" : "SELECT JSONExtractString(labels, 'k8s_cluster_name') AS k8s_cluster_name\nFROM signoz_metrics.distributed_time_series_v4_1day\nWHERE metric_name = 'k8s_node_cpu_time'\nGROUP BY k8s_cluster_name",
        "showALLOption" : false,
        "sort" : "DISABLED",
        "textboxValue" : "",
        "type" : "QUERY",
        "selectedValue" : "default",
        "allSelected" : false
      }
    }
  )
  version = "v4"
  widgets = jsonencode([
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "c5f29b09-8a63-44ba-825f-db91a3c79a54",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_cpu_utilization--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_cpu_utilization",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "14a0fa64",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "isJSON" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "6c8b0b46-1bd5-49aa-a4f9-0e4fd4636eaa",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node CPU usage",
      "yAxisUnit" : "none"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "230f3562-5ac1-4fec-946d-0dd21057f4b3",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_filesystem_available--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_filesystem_available",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "708a8633-bfaf-4bf4-b1e6-19ac4258c069",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node filesystem available",
      "yAxisUnit" : "bytes"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {
        "A" : "bytes",
        "B" : "bytes",
        "C" : "bytes"
      },
      "description" : "",
      "fillSpans" : false,
      "id" : "7403ba8f-36bf-4c31-8b91-a447a36eeca0",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "table",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_filesystem_usage--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_filesystem_usage",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "fccfd920",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "usage",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            },
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_filesystem_available--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_filesystem_available",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "B",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "edde9188",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "available",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "B",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            },
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_filesystem_capacity--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_filesystem_capacity",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "C",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "3a849415",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "capacity",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "C",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "4687e699-1594-47e3-b674-dde0f7de295a",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node filesystem capacity",
      "yAxisUnit" : "bytes"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "8722151e-7690-4152-98c3-f2cc0f741d50",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_filesystem_usage--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_filesystem_usage",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "985b3b6d",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "9138ac25-ad51-4de6-91f0-ab463542bee2",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node filesystem usage",
      "yAxisUnit" : "bytes"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "1c841c0b-be32-43ec-8bcb-bfd8a87edeef",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_available--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_memory_available",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "0b627b6e",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "4f7e31f7-0459-4aa6-aabc-26ad4ea07574",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node memory available",
      "yAxisUnit" : "bytes"
    },
    {
      "description" : "",
      "id" : "9d0f96dc-d744-4baa-9910-ac1aef63cc34",
      "isStacked" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_rss--float64----true",
                "isColumn" : true,
                "key" : "k8s_node_memory_rss",
                "type" : ""
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "991cf360",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "avg",
              "stepInterval" : 60,
              "timeAggregation" : "sum"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "ba8a15ef-ef6e-43e7-949a-5f59f763d4e7",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node memory rss",
      "yAxisUnit" : "bytes"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "75fdac11-19dd-472f-a155-63e4682b88df",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_usage--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_memory_usage",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "a41b8aba",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "72bb2f86-443f-4dbf-97b4-db4cfc083d11",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node memory usage",
      "yAxisUnit" : "bytes"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "960ff49f-d73b-49c2-ab4a-69df1e1abc51",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_working_set--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_memory_working_set",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "06ea5658",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "ef8b7768-8e5a-4db2-b084-f857ac58d289",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node memory working set",
      "yAxisUnit" : "bytes"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "1f8965fb-5ad1-4679-9d28-9bd31d4e4cac",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_network_errors--float64--Sum--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_network_errors",
                "type" : "Sum"
              },
              "aggregateOperator" : "rate",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "5ef4d1fe",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                },
                {
                  "dataType" : "string",
                  "id" : "interface--string--tag--false",
                  "isColumn" : false,
                  "key" : "interface",
                  "type" : "tag"
                },
                {
                  "dataType" : "string",
                  "id" : "direction--string--tag--false",
                  "isColumn" : false,
                  "key" : "direction",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}-{{interface}}-{{direction}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "rate"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "b997244c-136c-42d9-aeb6-2b3115202082",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node network errors",
      "yAxisUnit" : "none"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "2e864710-b418-4133-b248-2fa047e37fe3",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "graph",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_network_io--float64--Sum--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_network_io",
                "type" : "Sum"
              },
              "aggregateOperator" : "rate",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "4ced7960",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "direction--string--tag--false",
                  "isColumn" : false,
                  "key" : "direction",
                  "type" : "tag"
                },
                {
                  "dataType" : "string",
                  "id" : "interface--string--tag--false",
                  "isColumn" : false,
                  "key" : "interface",
                  "type" : "tag"
                },
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "{{k8s_node_name}}-{{interface}}-{{direction}}",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "sum",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "rate"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "88710842-8888-4225-b60f-eba516d16e07",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node network io",
      "yAxisUnit" : "bytes"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {},
      "description" : "",
      "fillSpans" : false,
      "id" : "f5174a53-c201-4e17-aff7-33b1402b0d7b",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "table",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_cpu_utilization--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_cpu_utilization",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "efcdf556",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "average cpu usage",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            },
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_allocatable_cpu--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_allocatable_cpu",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "B",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "66676d66",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "allocatable",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "B",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "611f9e7a-a319-45f4-971b-a1548489fa36",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Node CPU",
      "yAxisUnit" : "short"
    },
    {
      "bucketCount" : 30,
      "bucketWidth" : 0,
      "columnUnits" : {
        "A" : "bytes",
        "B" : "bytes",
        "C" : "bytes",
        "D" : "bytes",
        "E" : "bytes"
      },
      "description" : "",
      "fillSpans" : false,
      "id" : "4b9a4513-d7c8-4217-8d76-0714d96432e7",
      "isStacked" : false,
      "mergeAllActiveQueries" : false,
      "nullZeroValues" : "zero",
      "opacity" : "1",
      "panelTypes" : "table",
      "query" : {
        "builder" : {
          "queryData" : [
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_usage--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_memory_usage",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "A",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "7adc0e21",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "used",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "A",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            },
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_allocatable_memory--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_allocatable_memory",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "B",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "c721a7c9",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "allocatable",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "B",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            },
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_working_set--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_memory_working_set",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "C",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "65ec1def",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "working set",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "C",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            },
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_rss--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_memory_rss",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "D",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "a86caa2e",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "rss",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "D",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            },
            {
              "aggregateAttribute" : {
                "dataType" : "float64",
                "id" : "k8s_node_memory_available--float64--Gauge--true",
                "isColumn" : true,
                "isJSON" : false,
                "key" : "k8s_node_memory_available",
                "type" : "Gauge"
              },
              "aggregateOperator" : "avg",
              "dataSource" : "metrics",
              "disabled" : false,
              "expression" : "E",
              "filters" : {
                "items" : [
                  {
                    "id" : "4995e999",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_cluster_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_cluster_name",
                      "type" : "tag"
                    },
                    "op" : "=",
                    "value" : "{{.k8s_cluster_name}}"
                  },
                  {
                    "id" : "3632aed5",
                    "key" : {
                      "dataType" : "string",
                      "id" : "k8s_node_name--string--tag--false",
                      "isColumn" : false,
                      "key" : "k8s_node_name",
                      "type" : "tag"
                    },
                    "op" : "in",
                    "value" : [
                      "{{.k8s_node_name}}"
                    ]
                  }
                ],
                "op" : "AND"
              },
              "functions" : [],
              "groupBy" : [
                {
                  "dataType" : "string",
                  "id" : "k8s_node_name--string--tag--false",
                  "isColumn" : false,
                  "key" : "k8s_node_name",
                  "type" : "tag"
                }
              ],
              "having" : [],
              "legend" : "available",
              "limit" : null,
              "orderBy" : [],
              "queryName" : "E",
              "reduceTo" : "avg",
              "spaceAggregation" : "sum",
              "stepInterval" : 60,
              "timeAggregation" : "avg"
            }
          ],
          "queryFormulas" : []
        },
        "clickhouse_sql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "id" : "c1a1c09c-d98c-456d-b579-0983fbed6318",
        "promql" : [
          {
            "disabled" : false,
            "legend" : "",
            "name" : "A",
            "query" : ""
          }
        ],
        "queryType" : "builder"
      },
      "selectedLogFields" : [],
      "selectedTracesFields" : [],
      "softMax" : 0,
      "softMin" : 0,
      "stackedBarChart" : false,
      "thresholds" : [],
      "timePreferance" : "GLOBAL_TIME",
      "title" : "Memory usage",
      "yAxisUnit" : "bytes"
    }
  ])
}

output "dashboard_new" {
  value = signoz_dashboard.new_dashboard
}