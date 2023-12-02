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

const {
	useTextInput,
	useFileInput,
	useBoolInput,
	useFieldArrayInput
} = require("../lib/form");

const useFormSubmit = require("../lib/form/submit").default;
const { useWithFormContext, FormContext } = require("../lib/form/context");

const {
	TextInput,
	TextArea,
	FileInput,
	Checkbox
} = require("../components/form/inputs");

const FormWithData = require("../lib/form/form-with-data").default;
const FakeProfile = require("../components/fake-profile");
const MutationButton = require("../components/form/mutation-button");

const { useInstanceV1Query } = require("../lib/query");
const { useUpdateCredentialsMutation } = require("../lib/query/user");
const { useVerifyCredentialsQuery } = require("../lib/query/oauth");

module.exports = function UserProfile() {
	return (
		<FormWithData
			dataQuery={useVerifyCredentialsQuery}
			DataForm={UserProfileForm}
		/>
	);
};

function UserProfileForm({ data: profile }) {
	/*
		User profile update form keys
		- bool bot
		- bool locked
		- string display_name
		- string note
		- file avatar
		- file header
		- bool enable_rss
		- bool show_all_replies
		- bool noisy_mode
		- string custom_css (if enabled)
	*/

	const { data: instance } = useInstanceV1Query();
	const instanceConfig = React.useMemo(() => {
		return {
			allowCustomCSS: instance?.configuration?.accounts?.allow_custom_css === true,
			maxPinnedFields: instance?.configuration?.accounts?.max_profile_fields ?? 6
		};
	}, [instance]);

	const form = {
		avatar: useFileInput("avatar", { withPreview: true }),
		header: useFileInput("header", { withPreview: true }),
		displayName: useTextInput("display_name", { source: profile }),
		note: useTextInput("note", { source: profile, valueSelector: (p) => p.source?.note }),
		customCSS: useTextInput("custom_css", { source: profile, nosubmit: !instanceConfig.allowCustomCSS }),
		bot: useBoolInput("bot", { source: profile }),
		locked: useBoolInput("locked", { source: profile }),
		discoverable: useBoolInput("discoverable", { source: profile}),
		enableRSS: useBoolInput("enable_rss", { source: profile }),
		showAllReplies: useBoolInput("show_all_replies", { source: profile }),
		noisyMode: useBoolInput("noisy_mode", { source: profile }),
		fields: useFieldArrayInput("fields_attributes", {
			defaultValue: profile?.source?.fields,
			length: instanceConfig.maxPinnedFields
		}),
	};

	const [submitForm, result] = useFormSubmit(form, useUpdateCredentialsMutation(), {
		onFinish: () => {
			form.avatar.reset();
			form.header.reset();
		}
	});

	return (
		<form className="user-profile" onSubmit={submitForm}>
			<h1>Profile</h1>
			<div className="overview">
				<FakeProfile
					avatar={form.avatar.previewValue ?? profile.avatar}
					header={form.header.previewValue ?? profile.header}
					display_name={form.displayName.value ?? profile.username}
					username={profile.username}
					role={profile.role}
				/>
				<div className="files">
					<div>
						<FileInput
							label="Header"
							field={form.header}
							accept="image/*"
						/>
					</div>
					<div>
						<FileInput
							label="Avatar"
							field={form.avatar}
							accept="image/*"
						/>
					</div>
				</div>
			</div>

			<div className="form-section-docs">
				<h3>Basic Information</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#basic-information"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<TextInput
				field={form.displayName}
				label="Display name"
				placeholder="A GoToSocial user"
			/>
			<TextArea
				field={form.note}
				label="Bio"
				placeholder="Just trying out GoToSocial, my pronouns are they/them and I like sloths."
				rows={8}
			/>
			<b>Profile fields</b>
			<ProfileFields
				field={form.fields}
			/>

			<div className="form-section-docs">
				<h3>Visibility and privacy</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#visibility-and-privacy"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<Checkbox
				field={form.locked}
				label="Manually approve follow requests"
			/>
			<Checkbox
				field={form.discoverable}
				label="Mark account as discoverable by search engines and directories"
			/>
			<Checkbox
				field={form.enableRSS}
				label="Enable RSS feed of Public posts"
			/>
            <div className="form-section-docs">
				<h3>Timelines</h3>
                <a href="https://docs.gotosocial.org/en/latest/user_guide/settings/#timelines"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
                </a>
			</div>
			<Checkbox
				field={form.showAllReplies}
				label="Enable showing all follow's replies in home timeline"
			/>
			<Checkbox
				field={form.NoisyMode}
				label="Enable Noisy Mode"
			/>

			<div className="form-section-docs">
				<h3>Advanced</h3>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#advanced"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about these settings (opens in a new tab)
				</a>
			</div>
			<TextArea
				field={form.customCSS}
				label={`Custom CSS` + (!instanceConfig.allowCustomCSS ? ` (not enabled on this instance)` : ``)}
				className="monospace"
				rows={8}
				disabled={!instanceConfig.allowCustomCSS}
			/>
			<MutationButton label="Save profile info" result={result} />
		</form>
	);
}

function ProfileFields({ field: formField }) {
	return (
		<div className="fields">
			<FormContext.Provider value={formField.ctx}>
				{formField.value.map((data, i) => (
					<Field
						key={i}
						index={i}
						data={data}
					/>
				))}
			</FormContext.Provider>
		</div>
	);
}

function Field({ index, data }) {
	const form = useWithFormContext(index, {
		name: useTextInput("name", { defaultValue: data.name }),
		value: useTextInput("value", { defaultValue: data.value })
	});

	return (
		<div className="entry">
			<TextInput
				field={form.name}
				placeholder="Name"
			/>
			<TextInput
				field={form.value}
				placeholder="Value"
			/>
		</div>
	);
}
