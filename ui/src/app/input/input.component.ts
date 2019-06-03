import {Component, Input, OnInit} from '@angular/core';
import {GlobalPosition, InsidePlacement, Toppy, ToppyControl} from 'toppy';
import {APIService} from '../api.service';
import {OutputComponent} from '../output/output.component';

@Component({
  selector: 'app-input',
  templateUrl: './input.component.html',
  styleUrls: ['./input.component.scss']
})
export class InputComponent implements OnInit {
  @Input() featureLevel = 0;
  @Input() locked = true;

  popup: ToppyControl;
  popupOpened: boolean;
  format = 'video';
  quality = 'high';
  pageSize = 50;
  link = '';

  constructor(private toppy: Toppy,
              private api: APIService) { }

  ngOnInit() {
    this.popup = this.toppy
      .position(new GlobalPosition({
        placement: InsidePlacement.TOP,
        width: 'auto',
        height: 'auto',
        offset: 150
      }))
      .config({
        backdrop: true,
        closeOnEsc: true,
      })
      .content(OutputComponent)
      .create();

    this.popup.listen('t_close').subscribe(() => {
      this.popupOpened = false;
      this.link = '';
    });
  }

  submit() {
    this.api.createFeed({
      url: this.link,
      format: this.format,
      quality: this.quality,
      page_size: this.pageSize,
    }).subscribe(
      (resp) => {
        this.popup.content.props.address = resp.id;
        this.popup.open();
        this.popupOpened = true;
      },
      err => {
        alert(err);
      }
    );
  }

  allow600() {
    return !this.locked && this.featureLevel >= 2;
  }
}
