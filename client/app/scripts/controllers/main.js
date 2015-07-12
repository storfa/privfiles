'use strict';

/**
 * @ngdoc function
 * @name privfilesApp.controller:MainCtrl
 * @description
 * # MainCtrl
 * Controller of the privfilesApp
 */
angular.module('privfilesApp', ['angularFileUpload'])
  .controller('MainCtrl', function ($scope, $upload) {

    $scope.awesomeThings = [
      'HTML5 Boilerplate',
      'AngularJS',
      'Karma'
    ];
    //$scope.hasUploader = function(index) {
    //  return $scope.upload[index] !== null;
    //};

    $scope.abort = function(index) {
      $scope.upload[index].abort();
      $scope.upload[index] = null;
    };


    $scope.onFileSelect = function($files) {
      $scope.upload = [];
      $scope.progress = [];
      $scope.selectedFiles = [];
      $scope.uploadResult = null;

      var progressEvt = function(index, evt) {
        var percent = parseInt(100.0 * evt.loaded / evt.total, 10);
        console.log('percent: ' + percent);
        //$scope.progress.push(percent);
        $scope.progress[index] = percent;
      };

      var successEvt = function(index, data, status, headers, config) {
        // file is uploaded successfully
        console.log('\ndata:');
        console.log(data);

        console.log('\nstatus:');
        console.log(status);

        console.log('\nheaders:');
        console.log(headers);

        console.log('\nconfig:');
        console.log(config);
        $scope.uploadResult = data;
      };

      //$files: an array of files selected, each file has name, size, and type.
      for (var i = 0; i < $files.length; i++) {
        var file = $files[i];
        console.log(file);
        $scope.selectedFiles[i] = $files[i];
        $scope.upload[i] = $upload.upload({
          url: 'upload',
          method: 'POST',
          file: file//,
//          data: {
//            name: file.name,
//            size: file.size,
//            type: file.type
//          }
        }).progress(progressEvt.bind(this,i)).success(successEvt.bind(this,i));
      }
    };

  });
