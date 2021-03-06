import { Component, OnInit } from '@angular/core'
import { FormBuilder, FormGroup, Validators } from '@angular/forms'

import { MessageService } from 'primeng/api'

import { SettingsService } from '../backend/api/api'

@Component({
    selector: 'app-settings-page',
    templateUrl: './settings-page.component.html',
    styleUrls: ['./settings-page.component.sass'],
})
export class SettingsPageComponent implements OnInit {
    public settingsForm: FormGroup

    constructor(private fb: FormBuilder, private settingsApi: SettingsService, private msgSrv: MessageService) {
        this.settingsForm = this.fb.group({
            bind9_stats_puller_interval: ['', [Validators.required, Validators.min(0)]],
            grafana_url: [''],
            kea_hosts_puller_interval: ['', [Validators.required, Validators.min(0)]],
            kea_stats_puller_interval: ['', [Validators.required, Validators.min(0)]],
            kea_status_puller_interval: ['', [Validators.required, Validators.min(0)]],
            prometheus_url: [''],
        })
    }

    ngOnInit() {
        this.settingsApi.getSettings().subscribe(
            (data) => {
                const numericSettings = [
                    'bind9_stats_puller_interval',
                    'kea_hosts_puller_interval',
                    'kea_stats_puller_interval',
                    'kea_status_puller_interval',
                ]
                const stringSettings = ['grafana_url', 'prometheus_url']

                for (const s of numericSettings) {
                    if (data[s] === undefined) {
                        data[s] = 0
                    }
                }
                for (const s of stringSettings) {
                    if (data[s] === undefined) {
                        data[s] = ''
                    }
                }

                this.settingsForm.patchValue(data)
            },
            (err) => {
                let msg = err.statusText
                if (err.error && err.error.message) {
                    msg = err.error.message
                }
                this.msgSrv.add({
                    severity: 'error',
                    summary: 'Cannot get settings',
                    detail: 'Getting settings erred: ' + msg,
                    life: 10000,
                })
            }
        )
    }

    saveSettings() {
        if (!this.settingsForm.valid) {
            return
        }
        const settings = this.settingsForm.getRawValue()

        this.settingsApi.updateSettings(settings).subscribe(
            (data) => {
                this.msgSrv.add({
                    severity: 'success',
                    summary: 'Settings updated',
                    detail: 'Updating settings succeeded.',
                })
            },
            (err) => {
                let msg = err.statusText
                if (err.error && err.error.message) {
                    msg = err.error.message
                }
                this.msgSrv.add({
                    severity: 'error',
                    summary: 'Cannot get settings',
                    detail: 'Getting settings erred: ' + msg,
                    life: 10000,
                })
            }
        )
    }

    hasError(name, errType) {
        const setting = this.settingsForm.get(name)
        if (setting.errors && setting.errors[errType]) {
            return true
        }
        return false
    }
}
