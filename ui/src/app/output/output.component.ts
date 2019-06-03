import { Component, OnInit } from '@angular/core';
import {ToppyControl} from 'toppy';

@Component({
  selector: 'app-output',
  templateUrl: './output.component.html',
  styleUrls: ['./output.component.scss']
})
export class OutputComponent implements OnInit {
  address: string;
  canCopy: boolean;

  constructor(private overlay: ToppyControl) {
    this.address = overlay.content.props.address;
  }

  ngOnInit() {
    this.canCopy = document.queryCommandSupported('copy');
  }

  copyToClipboard(input) {
    input.select();
    document.execCommand('copy');
    input.setSelectionRange(0, 0);
  }

  close() {
    this.overlay.close();
  }
}
