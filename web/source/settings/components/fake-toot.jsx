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

const query = require("../lib/query");

module.exports = function FakeToot({ children }) {
	const { data: account = {
		avatar: "/assets/default_avatars/GoToSocial_icon1.png",
		display_name: "",
		username: ""
	} } = query.useVerifyCredentialsQuery();

	return (
		<article className="status expanded">
			<header className="status-header">
				<address>
					<a style={{margin: 0}}>
						<img className="avatar" src={account.avatar} alt="" />
						<dl className="author-strap">
							<dt className="sr-only">Display name</dt>
							<dd className="displayname text-cutoff">
								{account.display_name.trim().length > 0 ? account.display_name : account.username}
							</dd>
							<dt className="sr-only">Username</dt>
							<dd className="username text-cutoff">@{account.username}</dd>
						</dl>
					</a>
				</address>
			</header>
			<section className="status-body">
				<div className="text">
					<div className="content">
						{children}
					</div>
				</div>
			</section>
		</article>
	);
};