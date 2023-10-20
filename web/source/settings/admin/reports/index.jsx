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

const React = require("react");
const { Link, Switch, Route } = require("wouter");

const FormWithData = require("../../lib/form/form-with-data").default;

const ReportDetail = require("./detail");
const Username = require("./username");
const { useBaseUrl } = require("../../lib/navigation/util");
const { useListReportsQuery } = require("../../lib/query/admin/reports");

module.exports = function Reports({ baseUrl }) {
	return (
		<div className="reports">
			<Switch>
				<Route path={`${baseUrl}/:reportId`}>
					<ReportDetail />
				</Route>
				<ReportOverview />
			</Switch>
		</div>
	);
};

function ReportOverview({ }) {
	return (
		<>
			<h1>Reports</h1>
			<div>
				<p>
					Here you can view and resolve reports made to your instance, originating from local and remote users.
				</p>
			</div>
			<FormWithData
				dataQuery={useListReportsQuery}
				DataForm={ReportsList}
			/>
		</>
	);
}

function ReportsList({ data: reports }) {
	return (
		<div className="list">
			{reports.map((report) => (
				<ReportEntry key={report.id} report={report} />
			))}
		</div>
	);
}

function ReportEntry({ report }) {
	const baseUrl = useBaseUrl();
	const from = report.account;
	const target = report.target_account;

	let comment = report.comment.length > 200
		? report.comment.slice(0, 200) + "..."
		: report.comment;

	return (
		<Link to={`${baseUrl}/${report.id}`}>
			<a className={`report entry${report.action_taken ? " resolved" : ""}`}>
				<div className="byline">
					<div className="usernames">
						<Username user={from} link={false} /> reported <Username user={target} link={false} />
					</div>
					<h3 className="status">
						{report.action_taken ? "Resolved" : "Open"}
					</h3>
				</div>
				<div className="details">
					<b>Created: </b>
					<span>{new Date(report.created_at).toLocaleString()}</span>

					<b>Reason: </b>
					{comment.length > 0
						? <p>{comment}</p>
						: <i className="no-comment">none provided</i>
					}
				</div>
			</a>
		</Link>
	);
}