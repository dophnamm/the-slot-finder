<#-- The sign-in form: username and password, remember me and forgot password when the realm allows them. -->
<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('username','password') displayInfo=realm.password && realm.registrationAllowed && !registrationDisabled??; section>
    <#if section = "header">
        ${msg("loginAccountTitle")}
    <#elseif section = "subtitle">
        <p class="sf-subtitle">${msg("sfLoginSubtitle")}</p>
    <#elseif section = "form">
        <#if realm.password>
            <form id="kc-form-login" class="sf-form" onsubmit="login.disabled = true; return true;" action="${url.loginAction}" method="post">
                <#if !usernameHidden??>
                    <div class="${properties.kcFormGroupClass!}">
                        <label for="username" class="${properties.kcLabelClass!}"><#if !realm.loginWithEmailAllowed>${msg("username")}<#elseif !realm.registrationEmailAsUsername>${msg("usernameOrEmail")}<#else>${msg("email")}</#if></label>
                        <input tabindex="1" id="username" class="${properties.kcInputClass!}" name="username" value="${(login.username!'')}" type="text"
                               placeholder="${msg("sfUsernamePlaceholder")}" autofocus autocomplete="username" dir="ltr"
                               aria-invalid="<#if messagesPerField.existsError('username','password')>true</#if>" />
                    </div>
                </#if>
                <div class="${properties.kcFormGroupClass!}">
                    <div class="sf-label-row">
                        <label for="password" class="${properties.kcLabelClass!}">${msg("password")}</label>
                        <#if realm.resetPasswordAllowed>
                            <a tabindex="5" class="sf-link" href="${url.loginResetCredentialsUrl}">${msg("doForgotPassword")}</a>
                        </#if>
                    </div>
                    <input tabindex="2" id="password" class="${properties.kcInputClass!}" name="password" type="password" autocomplete="current-password" dir="ltr"
                           placeholder="${msg("sfPasswordPlaceholder")}"
                           aria-invalid="<#if messagesPerField.existsError('username','password')>true</#if>" />
                    <#if messagesPerField.existsError('username','password')>
                        <span id="input-error" class="${properties.kcInputErrorMessageClass!}" aria-live="polite">
                            ${kcSanitize(messagesPerField.getFirstError('username','password'))?no_esc}
                        </span>
                    </#if>
                </div>
                <#if realm.rememberMe && !usernameHidden??>
                    <label class="sf-checkbox">
                        <input tabindex="3" id="rememberMe" name="rememberMe" type="checkbox" <#if login.rememberMe??>checked</#if>> ${msg("rememberMe")}
                    </label>
                </#if>
                <input type="hidden" id="id-hidden-input" name="credentialId" <#if auth.selectedCredential?has_content>value="${auth.selectedCredential}"</#if>/>
                <input tabindex="4" class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!}" name="login" id="kc-login" type="submit" value="${msg("doLogIn")}"/>
            </form>
            <button type="button" id="sf-account-help-open" class="sf-link sf-link-button" aria-haspopup="dialog">${msg("sfAccountHelpLink")}</button>
            <dialog id="sf-account-help" class="sf-dialog" aria-labelledby="sf-account-help-title">
                <div class="sf-dialog-body">
                    <h2 id="sf-account-help-title" class="sf-dialog-title">${msg("sfAccountHelpTitle")}</h2>
                    <p class="sf-dialog-text">${msg("sfAccountHelpBody")}</p>
                    <div class="sf-actions">
                        <a class="sf-button sf-button-primary" href="${msg("sfAccountHelpManualUrl")}" target="_blank" rel="noopener noreferrer">${msg("sfAccountHelpOpenManual")}</a>
                        <form method="dialog" class="sf-actions">
                            <button class="sf-button sf-button-secondary" autofocus>${msg("sfAccountHelpClose")}</button>
                        </form>
                    </div>
                </div>
            </dialog>
            <script src="${url.resourcesPath}/js/account-help.js" type="module"></script>
        </#if>
    <#elseif section = "info">
        <#if realm.password && realm.registrationAllowed && !registrationDisabled??>
            <span>${msg("noAccount")} <a tabindex="6" class="sf-link" href="${url.registrationUrl}">${msg("doRegister")}</a></span>
        </#if>
    </#if>
</@layout.registrationLayout>
