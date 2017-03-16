
	var ColumnSchema = Backbone.Model.extend({
		defaults: {
			id: "column_" + new Date().getTime(),
		    type: 'unknown',
		    column_id: 'null',
		    attributes: {},
		    section: 'time_series_columns',
		    schema: {"type":"unknown","column_id":"null","attributes":{}},
		    classification: {}
		},
		initialize: function() {
			_.bindAll(this,'updateSchema');
			// console.log("Column Schema created");
		},
		updateSchema: function() {
			this.attributes.schema = {
				"type": this.attributes.type,
				"column_id": this.attributes.column_id,
				"attributes": this.attributes.attributes
			}
			// if ("selector" == this.attributes.schema.type) {
			// 	console.log(this.attributes.schema.attributes.values);
			// 	this.attributes.schema.attributes.values = this.attributes.schema.attributes.values.split(",");
			// }
		}
	});


	var ColumnSchemaCollection = Backbone.Collection.extend({
		model: ColumnSchema
	});


	var AppView = Backbone.View.extend({
	    el: "body",

	    initialize: function() {
	        _.bindAll(
	        	this, 
	        	'render', 
	        	'prepareUpload', 
	        	'uploadFiles', 
	        	'deleteColumn', 
	        	'deleteColumns',
	        	"createColumn",
	        	"downloadSchema",
	        	'updateColumnType', 
	        	'updateColumnSection', 
	        	'updateColumnName', 
	        	'updateColumnAttribute'
	        );
	        this.files;
	        this.columns = new ColumnSchemaCollection;
	    },

	    events: {
	        "change input[type=file]": "prepareUpload",
	        "submit form": "uploadFiles",
	        "change select.column_type": "updateColumnType",
	        "change select.column_section": "updateColumnSection",
	        "change input.column_name": "updateColumnName",
	        "change .column_attribute .form-control": "updateColumnAttribute",
	        "click .delete": "deleteColumn",
	        "click .create": "createColumn",
	        "click .download": "downloadSchema"
	    },

	    prepareUpload: function(e) {
	    	this.files = e.target.files;
			$.each(this.files, function(key, value) {
				$("#upload-file-info").html( value.name );
	        });
	    },

	    deleteColumn: function(e) {
	    	var self = this;
	    	var id = $(e.target).attr("model");
	    	if ($(e.target).hasClass("fa")) {
				id = $(e.target).parent().attr("model");
			}
	    	this.columns.each(function(model) {
	    		if (id == model.attributes.id) {
		  			swal({
			            title: "Delete column",
			            text: "Do you want to delete?",
			            type: "info",
			            showCancelButton: true,
			            closeOnConfirm: true
			        },function() {
			        	// swal("Success","Column was deleted", "success");
			    		$("#"+id).remove();
			    		self.columns.remove(model);
			    	});
			    	return;
	    		}
	    	});
	    },
	    
	    deleteColumns: function(e) {
			while (0 < this.columns.length) {
				var model = this.columns.models[0];
				$("#"+model.attributes.id).remove();
				this.columns.remove(model);
			}
		},

	    createColumn: function(e) {
			var column = new ColumnSchema();
			this.columns.add(column);
			this.render();
	    },

	    downloadSchema: function(e) {
	    	var column_ids = {};
	    	var datasource_schema = {"key":[],"static_columns":[],"time_series_columns":[]};
	    	$(".panel-danger").removeClass("panel-danger").addClass("panel-default");
	    	$(".form-group").removeClass("has-warning");
	    	var error = false;
	    	this.columns.each(function(model) {
	    		// schema error checks
	    		// check for duplicate column_id values
	    		if (column_ids.hasOwnProperty(model.attributes.column_id)) {
	    			swal("Error", "Duplicate column_id: " + model.attributes.column_id,"error");
	    			$("#"+model.attributes.id).removeClass("panel-default").addClass("panel-danger");
	    			$("#"+column_ids[model.attributes.column_id]).removeClass("panel-default").addClass("panel-danger");
	    			$("#"+model.attributes.id+" .form-group").addClass("has-warning");
	    			error = true;
	    			return;
	    		}
	    		// check for unknown
	    		if ("unknown" == model.attributes.type) {
					swal("Error", "Unknown column type: " + model.attributes.column_id,"error");
					$("#"+model.attributes.id).removeClass("panel-default").addClass("panel-danger");
					$("#"+model.attributes.id+" .form-group").addClass("has-warning");
					error = true;
					return;
	    		}
	    		// selector options
	    		if ("selector" == model.attributes.type) {
	    			if (0 == model.attributes.attributes.values.length ||
	    				1 == model.attributes.attributes.values.length && "" == model.attributes.attributes.values) {
	    				swal("Error", "Missing \"values\" in selector column: " + model.attributes.column_id,"error");
	    				$("#"+model.attributes.id).removeClass("panel-default").addClass("panel-danger");
	    				$("#"+model.attributes.id+" .form-group").addClass("has-warning");
	    				error = true;
	    				return;
	    			}
	    		}
	    		// range errors
	    		if ("fixed_point" == model.attributes.type || "integer" == model.attributes.type) {
	    			if (model.attributes.min_value > model.attributes.max_value) {
						swal("Error", "Range error in column: " + model.attributes.column_id, "error");
	    				$("#"+model.attributes.id).removeClass("panel-default").addClass("panel-danger");
	    				$("#"+model.attributes.id+" .form-group").addClass("has-warning");
	    				error = true;
	    				return;
	    			}
	    		}
	    		// add column to schema
	    		datasource_schema[model.attributes.section].push(model.attributes.schema);
	    		column_ids[model.attributes.column_id] = model.attributes.id;
	    	});
	    	
	    	// schema error checks
	    	if (error) {return;}
	    	if (0 == datasource_schema.key.length) {
	    		swal("Error", "Need at least one \"key\" column", "error");
	    		return;
	    	} else if (1 == datasource_schema.key.length) {
	    		datasource_schema.key = datasource_schema.key[0];
	    	}
	    	
	    	// clean unused column sections
	    	if (0 == datasource_schema.static_columns) {
	    		delete datasource_schema.static_columns;
	    	}
	    	if (0 == datasource_schema.time_series_columns) {
	    		delete datasource_schema.time_series_columns;
	    	}

	    	// create and download file
			var data = "text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(datasource_schema));
			$('<a id="downloadFile">download JSON</a>').appendTo("body");
			$("#downloadFile").attr("href", "data:" + data);
			$("#downloadFile").attr("download",  this.classification.id + '.json');
			$("#downloadFile")[0].click();
			$("#downloadFile").remove();

	    },

	    updateColumnType: function(e) {
	    	var self = this;
	    	var id = $(e.target).attr("model");
	    	this.columns.each(function(model) {
	    		if (id == model.attributes.id) {
	    			model.attributes.type = $(e.target).val();
	    			$("#"+id+" .column_attributes").html("")
	    			switch(model.attributes.type) {
					    case "date":
					    case "uuid":
						case "timestamp":
						case "unknown":
					    case "geographic_point":
					    	// $("#"+id+" .column_attributes").html("");
					    	model.attributes.attributes = {};
					        break;
						case "selector":
							model.attributes.attributes = {values:[]};
							break;
						case "varchar":
							model.attributes.attributes = {length:1};
							break;
						case "integer":
							if (!model.attributes.attributes.hasOwnProperty("min_value")) {
								model.attributes.attributes = {min_value:0,max_value:0};
							} else {
								delete model.attributes.attributes.precision;
							}
							break;
						case "fixed_point":
							if (!model.attributes.attributes.hasOwnProperty("min_value")) {
								model.attributes.attributes = {min_value:0,max_value:0,precision:1};
							} else {
								model.attributes.attributes.precision = 1;
							}
							break;

					    default:
					        console.log("error!");
					}
					// change gui
					$("#"+id+" .column_attributes").append(self.attributesHtml(model.attributes));
					// update model
	    			model.updateSchema();
	    			// console.log(model.attributes.schema);
					$("#"+model.attributes.id+" .column_schema").text(
		    			JSON.stringify(model.attributes.schema, null, 2)
		    		);
	    			return;
	    		}
	    	});
	    },

	    updateColumnSection: function(e) {
	    	id = $(e.target).attr("model");
	    	this.columns.each(function(model) {
	    		if (id == model.attributes.id) {
	    			model.attributes.section = $(e.target).val();
	    			model.updateSchema();
	    			console.log(model.attributes.schema);
	    			return;
	    		}
	    	});
	    },

	    updateColumnName: function(e) {
	    	id = $(e.target).attr("model");
	    	this.columns.each(function(model) {
	    		if (id == model.attributes.id) {
	    			model.attributes.column_id = $(e.target).val()
	    			model.updateSchema();
	    			console.log(model.attributes.schema);
	    			return;
	    		}
	    	});
	    },

	    updateColumnAttribute: function(e) {
	    	id = $(e.target).attr("model");
	    	this.columns.each(function(model) {
	    		if (id == model.attributes.id) {
	    			var attribute_name = $(e.target).attr("name");
	    			if ("number" == $(e.target).attr("type")) {
	    				model.attributes.attributes[attribute_name] = parseInt($(e.target).val());
	    			} else if ("values" ==  $(e.target).attr("name")) {
	    				// normalize text
	    				// remove \t \r \n
	    				// strip code tags
	    				// etc
	    				var text = $(e.target).val();
	    				var text = text.replace(", ", ",");
	    				var arr = text.split(",");
	    				var map = {};
	    				for (var i in arr) {
	    					map[arr[i]] = 0;
	    				}
	    				var v = Object.keys(map);
	    				model.attributes.attributes[attribute_name] = v;
	    				$(e.target).val(v);

	    			} else {
	    				model.attributes.attributes[attribute_name] = $(e.target).val();
	    			}
	    			model.updateSchema();
	    			// console.log(model.attributes.schema);
					$("#"+model.attributes.id+" .column_schema").text(
		    			JSON.stringify(model.attributes.schema, null, 2)
		    		);
	    			return;
	    		}
	    	});
	    },

	    uploadFiles: function(e) {
	    	var self = this;
			e.stopPropagation(); // Stop stuff happening
			e.preventDefault();  // Totally stop stuff happening

			if (!this.files) {
				swal("Error!", "Please provide a cvs file to upload", "error");
				return;
			}

	        // Create a formdata object and add the files
	        var data = new FormData();
	        var error = false;
	        $.each(this.files, function(key, value) {
	            data.append("uploadfile", value)
	            console.log(value.name);
	            // if (value.size > 21000000) {
	            if (value.size > 21000000000) {
	            	error = true;
	            	swal("Error", "File is to large","error")
	            	return;
	            }
	        });

	        if (error) {return;}

	        swal({
	            title: "Upload file and classify columns",
	            text: "Would you like to continue?",
	            type: "info",
	            showCancelButton: true,
	            closeOnConfirm: false,
	            showLoaderOnConfirm: true
	        }, function () {
				// remove columns
				self.deleteColumns();
				$('.uploadCsvFile').prop('disabled', true);
	            // Upload file request
	            $.ajax({
	                url: '/create_schema',
	                type: 'POST',
	                data: data,
	                cache: false,
	                dataType: 'json',
	                processData: false, // Don't process the files
	                contentType: false, // Set content type to false as jQuery will tell the server its a query string request
	                success: function(data, textStatus, jqXHR) {
	                    $('.uploadCsvFile').prop('disabled', false);
	                    if ("ok" == data.status) {
	                        swal("Success!", "Classification complete", "success");
	                        console.log(data);
	                        self.classification = data.data;
	                        for (var i in data.data.columns) {
	                        	var column = new ColumnSchema({
	                        		id: "column_" + i + "_" + new Date().getTime(),
								    type: data.data.columns[i].type,
								    column_id: data.data.columns[i].column_id,
								    attributes: data.data.columns[i].attributes,
								    section: "time_series_columns",
								    schema: data.data.columns[i]
	                        	});
	                        	self.columns.add(column);
	                        }
	                        self.render();
	                    } else {
	                        swal("Error!", JSON.stringify(data, null, 2), "error");
	                    }
	                },
	                error: function(jqXHR, textStatus, errorThrown){
	                    // Handle errors here
	                    $('.uploadCsvFile').prop('disabled', false);
	                    swal("Error!", textStatus, "error");
	                    $("#schema code").text(textStatus);
	                }
	            });
	        });

	    },

	    attributesHtml: function(column) {
    		if ("varchar" == column.type) {
				return [
					$("<div>").addClass("row").addClass("column_attribute").append(
						$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
							$("<label>").text("length: ")
						),
						$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
							$('<input/>').attr({ 
								type: 'number', 
								name: 'length', 
								value: column.attributes.length,
								model: column.id,
								min: 1
							}).addClass("form-control")
						)
					)
				]
    		}
    		else if ("selector" == column.type) {
    			return [
    				$("<div>").addClass("row").addClass("column_attribute").append(
						$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
							$("<label>").text("values: ")
						),
						$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
			    			$("<textarea>", {
			    				rows: "4",
			    				cols: "50",
			    				name: "values",
			    				model: column.id
			    			}).addClass("form-control").text(column.attributes.values)
		    			)
	    			)
    			];
    		}
    		else if ("integer" == column.type) {
    			return [
    				$("<div>").addClass("row").addClass("column_attribute").append(
						$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
							$("<label>").text("min_value: ")
						),
						$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
							$('<input/>').attr({ 
								type: 'number', 
								name: 'min_value', 
								value: column.attributes.min_value,
								model: column.id
							}).addClass("form-control")
						)
    				),
					$("<div>").addClass("row").addClass("column_attribute").append(
						$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
							$("<label>").text("max_value: ")
						),
						$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
							$('<input/>').attr({ 
								type: 'number', 
								name: 'max_value', 
								value: column.attributes.max_value,
								model: column.id
							}).addClass("form-control")
						)
    				)
    			];
    		}
    		else if ("fixed_point" == column.type) {
    			return [
    				$("<div>").addClass("row").addClass("column_attribute").append(
						$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
							$("<label>").text("min_value: ")
						),
						$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
							$('<input/>').attr({ 
								type: 'number', 
								name: 'min_value', 
								value: column.attributes.min_value,
								model: column.id
							}).addClass("form-control")
						)
    				),
					$("<div>").addClass("row").addClass("column_attribute").append(
						$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
							$("<label>").text("max_value: ")
						),
						$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
							$('<input/>').attr({ 
								type: 'number', 
								name: 'max_value', 
								value: column.attributes.max_value,
								model: column.id
							}).addClass("form-control")
						)
    				),
					$("<div>").addClass("row").addClass("column_attribute").append(
						$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
							$("<label>").text("precision: ")
						),
						$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
							$('<input/>').attr({ 
								type: 'number', 
								name: 'precision', 
								value: column.attributes.precision,
								model: column.id,
								min: 1
							}).addClass("form-control")
						)
    				)
    			];
    		}
	    },

	    columnHtml: function(column) {
	    	var self = this;
			var html = $("<div>").addClass("col-md-12").addClass("column").append(
				$("<div>").addClass("col-md-6").addClass("column").addClass("form-group").append(
					// $("<span>").addClass("glyphicon").addClass("glyphicon-warning-sign").addClass("form-control-feedback"),
					$("<div>").addClass("row").append(
						$("<div>").addClass("row").append(
							$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
								$("<label>").text("column_id: ")
							),
							$("<div>").addClass("col-md-9").addClass("col-xs-6").append(
								$('<input/>').attr({ 
									type: 'text', 
									name: 'column_id', 
									value: column.column_id,
									model: column.id
								}).addClass("column_name").addClass("form-control")
							)
						)
					),
			    	$("<div>").addClass("row").append(
						$("<div>").addClass("row").append(
							$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
								$("<label>").text("type: ")
							),
							$("<div>").addClass("col-md-9").addClass("col-xs-6").addClass("column").append(
								$("<select>", {model: column.id}).addClass("form-control").addClass("column_type").append(
									$("<option>").text("unknown"),
						    		$("<option>").text("selector"),
						    		$("<option>").text("varchar"),
						    		$("<option>").text("integer"),
						    		$("<option>").text("fixed_point"),
						    		$("<option>").text("geographic_point"),
						    		$("<option>").text("timestamp"),
						    		$("<option>").text("date"),
						    		$("<option>").text("uuid")
						    	).val(column.type)
							)
						)
				    ),
			    	$("<div>").addClass("row").append(
						$("<div>").addClass("row").append(
							$("<div>").addClass("col-md-3").addClass("col-xs-6").append(
								$("<label>").text("section: ")
							),
							$("<div>").addClass("col-md-9").addClass("col-xs-6").addClass("column").append(
								$("<select>", {model: column.id}).addClass("form-control").addClass("column_section").append(
									$("<option>").text("key"),
						    		$("<option>").text("static_columns"),
						    		$("<option>").text("time_series_columns")
						    	).val("time_series_columns")
							)
						)
				    ),
				    $("<div>").addClass("row").addClass("column_attributes").append(self.attributesHtml(column))
				),
				$("<div>").addClass("col-md-6").addClass("column").append(
					$("<div>").addClass("well").addClass("column_schema")
				)
		    );
		    return html;
		},

	    render: function() {
	    	var self = this;
	    	$(".panel-group").html('');
	    	// create column elements
	    	this.columns.each(function(model) {
	    		$(".panel-group").append(
	    			$("<div>",{id: model.attributes.id}).addClass("panel").addClass("panel-default").append(
	    				// $("<div>").addClass("panel-heading").text(model.attributes.id),
	    				$("<div>").addClass("panel-heading").addClass("panel-heading-sm").append(
	    					$("<div>").addClass("row").append(
							/*
		    					$("<div>").addClass("col-md-10").addClass("column").append(
		    						$("<label>").text(model.attributes.id)
		    					),
							*/			
								//$("<div>").addClass("col-md-2").addClass("column").append(
								$("<div>").addClass("col-md-12").addClass("column").append(
									$("<button>", {model:model.attributes.id}).addClass("btn").addClass("btn-danger").addClass("btn-xs").addClass("delete").addClass("right").append(
										$("<i>", {"aria-hidden":"true"}).addClass("fa").addClass("fa-times") //," Delete"
									)
								)
							)
	    				),
	    				$("<div>").addClass("panel-body").addClass("panel-body-sm").append(
			    			self.columnHtml(model.attributes)
	    				)
	    			)
	    		);
	    	});
	    	// display column schema
	    	this.columns.each(function(model) {
	    		// console.log($("#"+model.attributes.id+" .column_schema"));
	    		$("#"+model.attributes.id+" .column_schema").text(
	    			JSON.stringify(model.attributes.schema, null, 2)
	    		);
	    	});
	    }

	});


	var app = new AppView();
