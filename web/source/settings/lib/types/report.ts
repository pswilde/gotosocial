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

/**
 * Admin model of a report. Differs from the client
 * model, which contains less detailed information.
 */
export interface AdminReport {
    /**
	 * ID of the report.
	 */
	id: string;
	/**
	 * Whether an action has been taken by an admin in response to this report.
	 */
	action_taken: boolean;
	/**
	 * Time action was taken, if at all.
	 */
	action_taken_at?: string;
	/**
	 * Category under which this report was created.
	 */
	category: string;
	/**
	 * Comment submitted by the report creator.
	 */
	comment: string;
	/**
	 * Report was/should be federated to remote instance.
	 */
	forwarded: boolean;
	/**
	 * Time when the report was created.
	 */
	created_at: string;
	/**
	 * Time when the report was last updated.
	 */
	updated_at: string;
	/**
	 * Account that created the report.
	 * TODO: model this properly.
	 */
	account: Object;
	/**
	 * Reported account.
	 * TODO: model this properly.
	 */
	target_account: Object;
	/**
	 * Admin account assigned to handle this report, if any.
	 * TODO: model this properly.
	 */
	assigned_account?: Object;
	/**
	 * Admin account that has taken action on this report, if any.
	 * TODO: model this properly.
	 */
	action_taken_by_account?: Object;
	/**
	 * Statuses cited by this report, if any.
	 * TODO: model this properly.
	 */
	statuses: Object[];
	/**
	 * Rules broken according to the reporter, if any.
	 * TODO: model this properly.
	 */
	rules: Object[];
	/**
	 * Comment stored about what action (if any) was taken.
	 */
	action_taken_comment?: string;
}

/**
 * Parameters for POST to /api/v1/admin/reports/{id}/resolve.
 */
export interface AdminReportResolveParams {
    /**
	 * The ID of the report to resolve.
	 */
	id: string;
	/**
	 * Comment to store about what action (if any) was taken.
	 * Will be shown to the user who created the report (if local).
	 */
	action_taken_comment?: string;
}

/**
 * Parameters for GET to /api/v1/admin/reports.
 */
export interface AdminReportListParams {
	/**
	 * If set, show only resolved (true) or only unresolved (false) reports.
	 */
	resolved?: boolean;
	/**
	 * If set, show only reports created by the given account ID.
	 */
	account_id?: string;
	/**
	 * If set, show only reports that target the given account ID.
	 */
	target_account_id?: string;
	/**
	 * If set, show only reports older (ie., lower) than the given ID.
	 * Report with the given ID will not be included in response.
	 */
	max_id?: string;
	/**
	 * If set, show only reports newer (ie., higher) than the given ID.
	 * Report with the given ID will not be included in response.
	 */
	since_id?: string;
	/**
	 * If set, show only reports *immediately newer* than the given ID.
	 * Report with the given ID will not be included in response.
	 */
	min_id?: string;
	/**
	 * If set, limit returned reports to this number.
	 * Else, fall back to GtS API defaults.
	 */
	limit?: number;
}
