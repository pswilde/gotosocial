/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

import React from "react";

import { useEffect } from "react";

import { useExportDomainListMutation } from "../../lib/query/admin/domain-permissions/export";
import useFormSubmit from "../../lib/form/submit";

import {
	RadioGroup,
	TextArea,
	Select,
} from "../../components/form/inputs";

import MutationButton from "../../components/form/mutation-button";

import { Error } from "../../components/error";
import ExportFormatTable from "./export-format-table";

import type {
	FormSubmitFunction,
	FormSubmitResult,
	RadioFormInputHook,
	TextFormInputHook,
} from "../../lib/form/types";

export interface ImportExportFormProps {
	form: {
		domains: TextFormInputHook;
		exportType: TextFormInputHook;
		permType: RadioFormInputHook;
	};
	submitParse: FormSubmitFunction;
	parseResult: FormSubmitResult;
} 

export default function ImportExportForm({ form, submitParse, parseResult }: ImportExportFormProps) {
	const [submitExport, exportResult] = useFormSubmit(form, useExportDomainListMutation());

	function fileChanged(e) {
		const reader = new FileReader();
		reader.onload = function (read) {
			const res = read.target?.result;
			if (typeof res === "string") {
				form.domains.value = res;
				submitParse();
			}
		};
		reader.readAsText(e.target.files[0]);
	}

	useEffect(() => {
		if (exportResult.isSuccess) {
			form.domains.setter(exportResult.data);
		}
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [exportResult]);

	return (
		<>
			<h1>Import / Export domain permissions</h1>
			<p>This page can be used to import and export lists of domain permissions.</p>
			<p>Exports can be done in various formats, with varying functionality and support in other software.</p>
			<p>Imports will automatically detect what format is being processed.</p>
			<ExportFormatTable />
			<div className="import-export">
				<TextArea
					field={form.domains}
					label="Domains"
					placeholder={`google.com\nfacebook.com`}
					rows={8}
				/>

				<RadioGroup
					field={form.permType}
				/>

				<div className="button-grid">
					<MutationButton
						label="Import"
						type="button"
						onClick={() => submitParse()}
						result={parseResult}
						showError={false}
						disabled={false}
					/>
					<label className="button with-icon">
						<i className="fa fa-fw " aria-hidden="true" />
						Import file
						<input
							type="file"
							className="hidden"
							onChange={fileChanged}
							accept="application/json,text/plain,text/csv"
						/>
					</label>
					<b /> {/* grid filler */}
					<MutationButton
						label="Export"
						type="button"
						onClick={() => submitExport("export")}
						result={exportResult} showError={false}
						disabled={false}
					/>
					<MutationButton
						label="Export to file"
						wrapperClassName="export-file-button"
						type="button"
						onClick={() => submitExport("export-file")}
						result={exportResult}
						showError={false}
						disabled={false}
					/>
					<div className="export-file">
						<span>
							as
						</span>
						<Select
							field={form.exportType}
							options={<>
								<option value="plain">Text</option>
								<option value="json">JSON</option>
								<option value="csv">CSV</option>
							</>}
						/>
					</div>
				</div>

				{parseResult.error && <Error error={parseResult.error} />}
				{exportResult.error && <Error error={exportResult.error} />}
			</div>
		</>
	);
}
