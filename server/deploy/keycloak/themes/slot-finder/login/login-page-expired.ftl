<#-- Shown when a sign-in page is reused or left open too long: start the flow again, or continue it. -->
<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
        ${msg("pageExpiredTitle")}
    <#elseif section = "form">
        <p id="instruction1" class="instruction">${msg("sfPageExpiredMessage")}</p>
        <div class="sf-actions">
            <a id="loginRestartLink" class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!}" href="${url.loginRestartFlowUrl}">${msg("sfStartAgain")}</a>
            <a id="loginContinueLink" class="${properties.kcButtonClass!} ${properties.kcButtonDefaultClass!}" href="${url.loginAction}">${msg("sfContinue")}</a>
        </div>
    </#if>
</@layout.registrationLayout>
