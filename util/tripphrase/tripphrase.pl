#! /usr/bin/perl -w
#
#  Tripphrase generator
#    http://worrydream.com/tripphrase
#
#  by Bret Victor
#    http://worrydream.com
#
#  This software is licensed under the terms of the open source MIT license.
#    http://www.opensource.org/licenses/mit-license.php
#
#  The word lists are from Princeton's WordNet project.
#    http://wordnet.princeton.edu/
#

use strict;
use CGI qw(:standard escapeHTML);
use Digest::MD5 qw(md5_hex);

sub main {
    my $password = param('q') || $ARGV[0] || "";
		my $salt = param('q') || $ARGV[1] || "rabbit";
    my $digest = md5_hex("$salt$password");
    my @indexes = map { hex } ($digest =~ /..../g);
    
    my $template = templateForIndex(shift(@indexes));
    my @types = (split / /, $template);
    my @phraseWords = map { wordForIndexAndType(shift(@indexes), $_) } @types;
    my $phrase = join " ", @phraseWords;
    
    print "($phrase)";
}
    

#---------------------------------------------
#  templates

my @templates = (
    "verb article adj noun",
    "article adj adj noun",
    "article adv adj noun",
    "adv verb article noun",
);

sub templateForIndex {
    my ($index) = @_;

    my $wrappedIndex = $index % @templates;
    return $templates[$wrappedIndex];
}

my $wordsdir = "util/tripphrase";

#---------------------------------------------
#  words

my %wordsByType;

my @wordTypes = qw/noun verb adj adv article/;

my %wordTypes;
$wordTypes{$_} = 1 foreach (@wordTypes);

sub wordForIndexAndType {
    my ($index, $type) = @_;

    my $words = wordsForType($type);
    return $type unless $words;
    
    my $wrappedIndex = $index % @$words;
    my $word = $words->[$wrappedIndex];

    chomp($word);
    return $word;
}

sub isWordTypeValid {
    my ($type) = @_;
    return exists $wordTypes{$type};
}

sub wordsForType {
    my ($type) = @_;
    
    return $wordsByType{$type} if exists $wordsByType{$type};
    return "" unless isWordTypeValid($type);
    createWordListForType($type) unless -f "$wordsdir/$type.txt";

    open WORDS, "$wordsdir/$type.txt";
    my @words = <WORDS>;
    $wordsByType{$type} = \@words;
    close WORDS;

    return $wordsByType{$type};
}

sub createWordListForType {
    my ($type) = @_;
    
    open WORDS, ">$wordsdir/$type.txt";
    open INDEX, "index.$type";

    foreach (<INDEX>) {
        next unless /^[a-z]/;
        my ($word) = /^(\S+)/;
        next if length($word) < 2;
        next if $word =~ /_/;
        
        print WORDS "$word\n";
    }
    
    close INDEX;
    close WORDS;
}


#---------------------------------------------
#  go

main();
