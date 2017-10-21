var gulp = require('gulp'),
    del = require('del'),
    path = require('path'),
    uglify = require('gulp-uglify'),
    rev = require('gulp-rev'),
    revreplace = require('gulp-rev-replace'),
    cleancss = require('gulp-clean-css'),
    autoprefixer = require('gulp-autoprefixer'),
    size = require('gulp-size'),
    gulpif = require('gulp-if'),
    imagemin = require('gulp-imagemin');

abs = path.join(process.cwd(), 'assets');

gulp.task('clean', function () {
    return del(['./dist/**/*'])
});

// Minify images and output to ./dist folder
gulp.task('img', ['clean'], function() {
    return gulp.src('./assets/**/*.{png,ico}')
        .pipe(imagemin())
        .pipe(size())
        .pipe(gulp.dest('./dist'))
});

// Minify scripts, build manifest.json and output to ./dist folder
gulp.task('js+css', ['clean', 'img'], function() {
    return gulp.src(['./assets/js/**/*.js', './assets/css/**/*.css'], {base: abs})
        .pipe(gulpif(/js$/, uglify()))
        .pipe(gulpif(/css$/, autoprefixer()))
        .pipe(gulpif(/css$/, cleancss()))
        .pipe(rev())
        .pipe(size())
        .pipe(gulp.dest('./dist'))
        .pipe(rev.manifest('manifest.json', {merge: true}))
        .pipe(gulp.dest('./dist'));
});

// Rewrite occurrences of scripts in template files
gulp.task('patch', ['js+css'], function() {
    var manifest = gulp.src('./dist/manifest.json');
    return gulp.src('./templates/index.html')
        .pipe(revreplace({manifest: manifest}))
        .pipe(gulp.dest('./templates/'))
});

gulp.task('default', ['js+css']);